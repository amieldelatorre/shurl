package internal

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/db"
	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
)

type App struct {
	Server    *http.Server
	Logger    utils.CustomJsonLogger
	Config    *config.Config
	DbContext db.DbContext
}

func NewApp(configFilePath string) App {
	logger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
	ctx := context.Background()

	config, err := config.LoadConfig(configFilePath, ctx, logger)
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}

	baseUrl := getBaseUrlString(config.Server.HttpsEnabled, strings.TrimSpace(config.Server.Domain), strings.TrimSpace(config.Server.Port))

	dbContext := db.GetDatabaseContext(ctx, *config, logger)

	mux := http.NewServeMux()

	middleware := handlers.NewMiddleware(logger, *config)
	apiHandler := handlers.NewApiShortUrlHandler(logger, dbContext, baseUrl)
	apiUserHandler := handlers.NewApiUserHandler(logger, dbContext)
	redirectionHandler := handlers.NewRedirectionHandler(logger, dbContext)

	templateHandler := handlers.NewTemplateHandler(logger, baseUrl)

	RegisterRoutes(logger, ctx, mux, middleware, apiHandler, apiUserHandler, redirectionHandler, templateHandler)

	app := App{
		Config: config,
		Logger: logger,
		Server: &http.Server{
			Addr:    ":" + config.Server.Port,
			Handler: mux,
		},
		DbContext: dbContext,
	}
	return app
}

func (a *App) Exit() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	a.Logger.Info(ctx, "Exiting application...")

	a.Logger.Info(ctx, "Shutting down server")
	err := a.Server.Shutdown(ctx)
	if err != nil {
		a.Logger.ErrorExit(ctx, "Error shutting down server", "error", err)
	}

	a.Logger.Info(ctx, "Application has been shutdown, bye bye !")
}

func (a *App) Run() {
	ctx := context.Background()
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		a.Logger.Info(ctx, "Attempting to start the server...")
		a.Logger.Info(ctx, fmt.Sprintf("Starting server on port %s", a.Server.Addr))
		err := a.Server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.ErrorExit(ctx, "Something went wrong with the server", "error", err)
		}
	}()

	sig := <-stopChan

	a.Logger.Info(ctx, fmt.Sprintf("Received signal '%+v', attempting to shutdown", sig))
	a.Exit()
}

func getBaseUrlString(httpsEnabled bool, domain string, port string) string {
	protocol := "http"
	if httpsEnabled {
		protocol = "https"
	}

	baseUrl := fmt.Sprintf("%s://%s", protocol, domain)

	isNotStandardHttp := !httpsEnabled && port != "80"
	isNotStandardHttps := httpsEnabled && port != "443"
	if isNotStandardHttp || isNotStandardHttps {
		baseUrl += fmt.Sprintf(":%s", port)
	}

	return baseUrl
}
