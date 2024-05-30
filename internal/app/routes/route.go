package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/Erlast/short-url.git/internal/app/middlewares"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func NewRouter(store storages.URLStorage, conf *config.Cfg, logger *zap.SugaredLogger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(func(h http.Handler) http.Handler {
		return middlewares.WithLogging(h, logger)
	})
	r.Use(func(h http.Handler) http.Handler {
		return middlewares.GzipMiddleware(h, logger)
	})

	handleGet := func(res http.ResponseWriter, req *http.Request) {
		handlers.GetHandler(res, req, store)
	}

	r.Get("/{id}", handleGet)

	handlePost := func(res http.ResponseWriter, req *http.Request) {
		handlers.PostHandler(res, req, store, conf)
	}

	r.Post("/", handlePost)

	handleShorten := func(res http.ResponseWriter, req *http.Request) {
		handlers.PostShortenHandler(res, req, store, conf)
	}

	r.Post("/api/shorten", handleShorten)

	r.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		handlers.GetPingHandler(res, req, conf)
	})

	return r
}
