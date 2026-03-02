package internal

import (
	"context"
	"embed"
	"io/fs"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/handlers"
	"github.com/amieldelatorre/shurl/internal/utils"
)

//go:embed static
var embedHtmlStatic embed.FS

func RegisterRoutes(
	logger utils.CustomJsonLogger,
	ctx context.Context,
	mux *http.ServeMux,
	m handlers.Middleware,
	apiShortUrlHandler handlers.ApiShortUrlHandler,
	apiUserHandler handlers.ApiUserHandler,
	authHandler handlers.ApiAuthHandler,
	apiHealthHandler handlers.ApiHealthHandler,
	redirectionHandler handlers.RedirectionHandler,
	templateHandler handlers.TemplateHandler,
) {
	redirection := m.RecoverPanic(m.AddRequestId(http.HandlerFunc(redirectionHandler.Redirect)))
	postShortUrl := m.RecoverPanic(m.AddRequestId(m.LoginRequiredOrAllowAnonymous(m.IdempotencyKeyRequired(http.HandlerFunc(apiShortUrlHandler.PostShortUrl)))))
	postUser := m.RecoverPanic(m.AddRequestId(m.AllowRegistration(m.IdempotencyKeyRequired(http.HandlerFunc(apiUserHandler.PostUser)))))
	getIndexJs := m.RecoverPanic(m.AddRequestId(http.HandlerFunc(templateHandler.GetIndexJs)))
	login := m.RecoverPanic(m.AddRequestId(m.AllowLogin(http.HandlerFunc(authHandler.Login))))
	logout := m.RecoverPanic(m.AddRequestId(http.HandlerFunc(authHandler.Logout)))
	validate := m.RecoverPanic(m.AddRequestId(m.LoginRequired(http.HandlerFunc(authHandler.Validate))))
	healthCheck := m.RecoverPanic(m.AddRequestId(http.HandlerFunc(apiHealthHandler.HealthCheck)))

	htmlSubFs, err := fs.Sub(embedHtmlStatic, "static")
	if err != nil {
		logger.ErrorExit(ctx, err.Error())
	}
	fileServer := http.FileServer(http.FS(htmlSubFs))

	mux.Handle("POST /api/v1/shorturl", postShortUrl)
	mux.Handle("POST /api/v1/user", postUser)
	mux.Handle("POST /api/v1/auth/login", login)
	mux.Handle("POST /api/v1/auth/logout", logout)
	mux.Handle("GET /api/v1/auth/validate", validate)
	mux.Handle("GET /api/v1/health", healthCheck)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	/*
	* In this section, even if there is an example.html in the static/ dir, it will still end up at the redirection path.
	* If there are more paths needed in the future, like login.html, it can be served on
	* "/_/" path with an http.StripPrefix("/_/") and point it to the file server again.
	 */
	mux.Handle("GET /", fileServer)
	mux.Handle("GET /_/", http.StripPrefix("/_/", fileServer))
	mux.Handle("GET /_/shared.js", http.StripPrefix("/_/", getIndexJs))
	mux.Handle("GET /{slug}", redirection)
}
