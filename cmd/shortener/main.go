package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/config"
	"github.com/Erlast/short-url.git/internal/handlers"
	"github.com/Erlast/short-url.git/internal/storages"
)

func main() {
	conf := config.ParseFlags()

	store := storages.Init(make(map[string]string))

	handlers.Init(store, conf)

	r := chi.NewRouter()

	r.Get("/{id}", handlers.GetHandler)

	r.Post("/", handlers.PostHandler)

	err := http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}
