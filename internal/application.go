package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/db/valkey_cache"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
	"github.com/amieldelatorre/shurl/internal/workers"
)

type App struct {
	Server       *http.Server
	Logger       utils.CustomJsonLogger
	Config       *config.Config
	DbContext    db.DbContext
	CacheContext *db.DbContext
	baseUrl      string
}

func NewApp(configFilePath string) App {
	tempLogger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
	ctx := context.Background()

	config, err := config.LoadConfig(configFilePath)
	if err != nil {
		tempLogger.ErrorExit(ctx, err.Error())
	}

	logger := utils.NewCustomJsonLogger(os.Stdout, config.Log.SlogLevel)

	if config.Server.HttpsEnabled {
		handlers.CookieAccessTokenName = "__Host-" + handlers.CookieAccessTokenName
	}

	baseUrl := getBaseUrlString(config.Server.HttpsEnabled, config.Server.Domain, config.Server.Port, config.Server.AppendPort)

	dbContext := db.GetDatabaseContext(ctx, *config, logger, false)
	actualDbContext := dbContext
	var cacheContext db.DbContext
	if config.Cache.Enabled != nil && *config.Cache.Enabled {
		cacheContext = valkey_cache.GetDatabaseContext(ctx, *config, logger, dbContext)
		dbContext = cacheContext
	}

	mux := http.NewServeMux()

	middleware := handlers.NewMiddleware(logger, *config)
	apiShortUrlHandler := handlers.NewApiShortUrlHandler(logger, dbContext, baseUrl)
	apiUserHandler := handlers.NewApiUserHandler(logger, dbContext)
	apiAuthHandler, err := handlers.NewApiAuthHandler(logger, *config, dbContext)
	apiHealthHandler := handlers.NewApiHealthHandler(logger, *config, actualDbContext, cacheContext)
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
		DbContext:    actualDbContext,
		CacheContext: &cacheContext,
		baseUrl:      baseUrl,
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
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.Logger.Info(ctx, "Attempting to start the server...")
		a.Logger.Info(ctx, fmt.Sprintf("Starting server on port %s", a.Server.Addr))
		a.Logger.Info(ctx, fmt.Sprintf("Server will be available on %s", a.baseUrl))
		errChan <- a.Server.ListenAndServe()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		workers.IdempotencyKeyCleanupWorker(ctx, a.Logger, a.Config.IdempotencyKeyCleanupWorker.IntervalSeconds, a.DbContext, a.Config.IdempotencyKeyCleanupWorker.ErrorsFatal)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		workers.ShortUrlCleanupWorker(ctx, a.Logger, a.Config.ShortUrlCleanupWorker.IntervalSeconds, a.DbContext, a.Config.ShortUrlCleanupWorker.ErrorsFatal)
	}()

	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			stop()
			a.Logger.Error(context.Background(), "Something went wrong with the server", "error", err)
		}
	case <-ctx.Done():
		a.Logger.Info(context.Background(), "Received signal, attempting to shutdown")

	}

	exitCtx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	a.Exit(exitCtx)
	wg.Wait()
	fmt.Println("All workers exited")
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
