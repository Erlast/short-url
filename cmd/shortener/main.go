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

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// main настройка приложения.
//
//go:generate go run ./../version/version.go
func main() {
	// Вспомогательная функция для профилирования
	go func() {
		log.Println(http.ListenAndServe("localhost:8070", nil))
	}()

	// Задаем конфигурацию сервера
	conf := config.ParseFlags()

	// Контекст
	ctx := context.Background()

	// Иницциализация логгирования
	newLogger, err := logger.NewLogger("info")
	if err != nil {
		log.Fatal("Running logger fail")
	}

	// Инициализация хранилища
	store, err := storages.NewStorage(ctx, conf, newLogger)
	if err != nil {
		newLogger.Fatalf("Unable to create storage %v: ", err)
	}

	// Запуск компонента удаления записей, которые ранее были мягко удалены
	go components.DeleteSoftDeletedRecords(ctx, store)

	// Инициализация роутов
	r := routes.NewRouter(ctx, store, conf, newLogger)

	// Вывод информации в лог о старте сервера
	newLogger.Info("Running server address ", conf.FlagRunAddr)
	newLogger.Infof("Build version: %s\n", buildVersion)
	newLogger.Infof("Build date: %s\n", buildDate)
	newLogger.Infof("Build commit: %s\n", buildCommit)

	// Запуск сервера
	err = http.ListenAndServe(conf.FlagRunAddr, r)

	if err != nil {
		newLogger.Fatal("Running server fail")
	}
}
