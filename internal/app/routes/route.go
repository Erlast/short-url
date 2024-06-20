package routes

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/Erlast/short-url.git/internal/app/middlewares"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func NewRouter(ctx context.Context, store storages.URLStorage, conf *config.Cfg, logger *zap.SugaredLogger) *chi.Mux {
	r := chi.NewRouter()

	currentUser := &storages.CurrentUser{}

	r.Use(func(h http.Handler) http.Handler { return middlewares.AuthMiddleware(h, logger, currentUser) })
	r.Use(func(h http.Handler) http.Handler {
		return middlewares.WithLogging(h, logger)
	})
	r.Use(func(h http.Handler) http.Handler {
		return middlewares.GzipMiddleware(h, logger)
	})

	r.Get("/{id}", func(res http.ResponseWriter, req *http.Request) {
		handlers.GetHandler(ctx, res, req, store)
	})

	r.Post("/", func(res http.ResponseWriter, req *http.Request) {
		handlers.PostHandler(ctx, res, req, store, conf, logger, currentUser)
	})

	r.Post("/api/shorten", func(res http.ResponseWriter, req *http.Request) {
		handlers.PostShortenHandler(ctx, res, req, store, conf, logger, currentUser)
	})

	r.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		handlers.GetPingHandler(ctx, res, store, logger)
	})

	r.Post("/api/shorten/batch", func(res http.ResponseWriter, req *http.Request) {
		handlers.BatchShortenHandler(ctx, res, req, store, conf, logger, currentUser)
	})

	r.Route("/api/user/urls", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler { return middlewares.CheckAuthMiddleware(h, logger) })
		r.Get("/", func(res http.ResponseWriter, req *http.Request) {
			handlers.GetUserUrls(ctx, res, store, conf, logger, currentUser)
		})
		r.Delete("/", func(res http.ResponseWriter, req *http.Request) {
			handlers.DeleteUserUrls(ctx, res, req, store, logger, currentUser)
		})
	})

	return r
}
