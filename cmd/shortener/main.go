package main

import (
	"log"
	"net/http"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/middlewares"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func main() {
	conf := config.ParseFlags()

	newLogger, err := logger.NewLogger("info")

	if err != nil {
		log.Fatal("Running logger fail")
	}

	store := storages.NewStorage(conf)

	r := routes.NewRouter(store, conf)

	newLogger.Info("Running server address ", conf.FlagRunAddr)

	err = http.ListenAndServe(conf.FlagRunAddr, middlewares.WithLogging(middlewares.GzipMiddleware(r)))

	if err != nil {
		newLogger.Fatal("Running server fail")
	}
}
