package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Erlast/short-url.git/internal/app"
	"github.com/Erlast/short-url.git/internal/config"
)

func main() {
	conf := config.ParseFlags()

	app.Init(app.Settings{
		Storage: make(map[string]string),
		Conf:    conf,
	})

	r := chi.NewRouter()

	r.Get("/{id}", app.GetHandler)

	r.Post("/", app.PostHandler)

	err := http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}
