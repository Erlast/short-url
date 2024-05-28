package main

import (
	"log"
	"net/http"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func main() {
	conf := config.ParseFlags()

	newLogger, err := logger.NewLogger("info")

	if err != nil {
		log.Fatal("Running logger fail")
	}

	store, err := storages.NewStorage(conf, newLogger)
	if err != nil {
		newLogger.Fatalf("Unable to create storage %v: ", err)
	}

	r := routes.NewRouter(store, conf, newLogger)

	newLogger.Info("Running server address ", conf.FlagRunAddr)

	err = http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		newLogger.Fatal("Running server fail")
	}
}
