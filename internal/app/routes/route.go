package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/handlers"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func Init(store *storages.Storage, conf *config.Cfg) *chi.Mux {
	r := chi.NewRouter()

	handleGet := func(res http.ResponseWriter, req *http.Request) {
		handlers.GetHandler(res, req, store)
	}

	r.Get("/{id}", handleGet)

	handlePost := func(res http.ResponseWriter, req *http.Request) {
		handlers.PostHandler(res, req, store, conf)
	}

	r.Post("/", handlePost)

	return r
}
