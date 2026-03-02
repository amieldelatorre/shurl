package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/amieldelatorre/shurl/internal/workers"
)

type App struct {
	Server    *http.Server
	Logger    utils.CustomJsonLogger
	Config    *config.Config
	DbContext db.DbContext
	baseUrl   string
}

func NewApp(configFilePath string) App {
	logger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
	ctx := context.Background()

	config, err := config.LoadConfig(configFilePath, ctx, logger)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	if config.Server.HttpsEnabled {
		handlers.CookieAccessTokenName = "__Host-" + handlers.CookieAccessTokenName
	}

	baseUrl := getBaseUrlString(config.Server.HttpsEnabled, config.Server.Domain, config.Server.Port, config.Server.AppendPort)

	dbContext := db.GetDatabaseContext(ctx, *config, logger)

	mux := http.NewServeMux()

	middleware := handlers.NewMiddleware(logger, *config)
	apiShortUrlHandler := handlers.NewApiShortUrlHandler(logger, dbContext, baseUrl)
	apiUserHandler := handlers.NewApiUserHandler(logger, dbContext)
	apiAuthHandler, err := handlers.NewApiAuthHandler(logger, *config, dbContext)
	apiHealthHandler := handlers.NewApiHealthHandler(logger, *config, dbContext)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	redirectionHandler := handlers.NewRedirectionHandler(logger, dbContext)
	templateHandler := handlers.NewTemplateHandler(logger, baseUrl, *config)

	RegisterRoutes(logger, ctx, mux, middleware, apiShortUrlHandler, apiUserHandler, apiAuthHandler, apiHealthHandler, redirectionHandler, templateHandler)

	app := App{
		Config: config,
		Logger: logger,
		Server: &http.Server{
			Addr:    ":" + config.Server.Port,
			Handler: mux,
		},
		DbContext: dbContext,
		baseUrl:   baseUrl,
	}
	return app
}

func (a *App) Exit(ctx context.Context) {
	a.Logger.Info(ctx, "Exiting application...")

	a.Logger.Info(ctx, "Shutting down server")
	err := a.Server.Shutdown(ctx)
	if err != nil {
		a.Logger.ErrorExit(ctx, "Error shutting down server", "error", err)
	}

	a.Logger.Info(ctx, "Application has been shutdown, bye bye !")
}

func (a *App) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	errChan := make(chan error, 1)

	go func() {
		a.Logger.Info(ctx, "Attempting to start the server...")
		a.Logger.Info(ctx, fmt.Sprintf("Starting server on port %s", a.Server.Addr))
		a.Logger.Info(ctx, fmt.Sprintf("Server will be available on %s", a.baseUrl))
		errChan <- a.Server.ListenAndServe()
	}()

	go func() {
		workers.IdempotencyKeyCleanupWorker(ctx, a.Logger, a.Config.IdempotencyKeyCleanupWorker.IntervalSeconds, a.DbContext, a.Config.IdempotencyKeyCleanupWorker.ErrorsFatal)
	}()

	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.ErrorExit(context.Background(), "Something went wrong with the server", "error", err)
		}
	case <-ctx.Done():
		a.Logger.Info(context.Background(), "Received signal, attempting to shutdown")

	}

	exitCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	a.Exit(exitCtx)
}

func getBaseUrlString(httpsEnabled bool, domain string, port string, appendPort bool) string {
	protocol := "http"
	if httpsEnabled {
		protocol = "https"
	}

	baseUrl := fmt.Sprintf("%s://%s", protocol, domain)

	isNotStandardHttp := !httpsEnabled && port != "80"
	isNotStandardHttps := httpsEnabled && port != "443"
	if (isNotStandardHttp || isNotStandardHttps) && appendPort {
		baseUrl += fmt.Sprintf(":%s", port)
	}

	return baseUrl
}
