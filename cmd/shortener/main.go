package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Erlast/short-url.git/internal/app/components"
	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:8070", nil))
	}()

	conf := config.ParseFlags()
	ctx := context.Background()

	newLogger, err := logger.NewLogger("info")

	if err != nil {
		log.Fatal("Running logger fail")
	}

	store, err := storages.NewStorage(ctx, conf, newLogger)
	if err != nil {
		newLogger.Fatalf("Unable to create storage %v: ", err)
	}

	go components.DeleteSoftDeletedRecords(ctx, store)

	r := routes.NewRouter(ctx, store, conf, newLogger)

	newLogger.Info("Running server address ", conf.FlagRunAddr)

	err = http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		newLogger.Fatal("Running server fail")
	}
}
