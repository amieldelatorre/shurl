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
)

type App struct {
	Server    *http.Server
	Logger    utils.CustomJsonLogger
	Config    *config.Config
	DbContext db.DbContext
}

func NewApp(configFilePath string) App {
	logger := utils.NewCustomJsonLogger(os.Stdout, slog.LevelDebug)
	config, err := config.LoadConfig(configFilePath)
	if err != nil {
		ctx := context.Background()
		logger.ErrorExit(ctx, err.Error())
	}

	ctx := context.Background()
	dbContext := db.GetDatabaseContext(ctx, *config, logger)

	mux := http.NewServeMux()

	middleware := handlers.NewMiddleware(logger)
	apiHandler := handlers.NewApiHandler(logger, dbContext, *config)
	redirectionHandler := handlers.NewRedirectionHandler(logger, dbContext)

	RegisterRoutes(logger, ctx, mux, middleware, apiHandler, redirectionHandler)

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
