package main

import (
	"context"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func TestMain(m *testing.M) {
	conf := config.Cfg{
		FlagRunAddr: ":8082",
		FileStorage: "/tmp/short-test.json",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Инициализация логгера
	newLogger, err := logger.NewLogger("info")
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	ctrl := gomock.NewController(nil)
	defer ctrl.Finish()

	// Создание мока хранилища
	store := storages.NewMockURLStorage(ctrl)

	// Инициализация роутера
	r := routes.NewRouter(ctx, store, &conf, newLogger)

	// Запуск сервера в отдельной горутине
	server := &http.Server{
		Addr:    conf.FlagRunAddr,
		Handler: r,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			newLogger.Fatalf("server failed: %v", err)
		}
	}()

	// Ожидание, чтобы сервер успел запуститься
	time.Sleep(1 * time.Second)

	// Проверка ответа от сервера
	resp, err := http.Get("http://localhost" + conf.FlagRunAddr)
	if err != nil {
		panic("failed to get response from server: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		panic("expected status OK from server, got " + http.StatusText(resp.StatusCode))
	}

	if err := server.Shutdown(ctx); err != nil {
		panic("server Shutdown Failed:" + err.Error())
	}

	os.Exit(m.Run())
}

func BenchmarkRandomString(b *testing.B) {
	tests := []int{7, 14, 28, 56, 112}

	for _, n := range tests {
		b.Run("Length_"+string(rune(n)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				helpers.RandomString(n)
			}
		})
	}
}
