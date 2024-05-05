package main

import (
	"log"
	"net/http"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func main() {
	conf := config.ParseFlags()

	store := storages.Init()

	r := routes.Init(store, conf)

	err := http.ListenAndServe(conf.GetRunAddr(), r)

	if err != nil {
		log.Fatal(err)
	}
}
