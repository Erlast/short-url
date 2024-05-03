package main

import (
	"github.com/Erlast/short-url.git/internal/config"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"

	"github.com/Erlast/short-url.git/internal/app"
)

func main() {

	conf, err := config.ParseFlags()

	if err != nil {
		log.Fatal(err)
	}

	app.Init(app.Settings{
		Storage: make(map[string]string),
		Conf:    conf,
	})

	r := chi.NewRouter()

	r.Get("/{id}", app.GetHandler)

	r.Post("/", app.PostHandler)

	err = http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		log.Fatal(err)
	}
}
