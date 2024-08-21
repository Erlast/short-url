package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/Erlast/short-url.git/internal/app/components"
	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/routes"
	"github.com/Erlast/short-url.git/internal/app/storages"
	"github.com/stretchr/testify/assert"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

func TestMainFunction(t *testing.T) {
	conf := config.Cfg{
		FlagRunAddr: ":8080",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newLogger, err := logger.NewLogger("info")
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	store, err := storages.NewStorage(ctx, &conf, newLogger)
	if err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	go components.DeleteSoftDeletedRecords(ctx, store)

	r := routes.NewRouter(ctx, store, &conf, newLogger)

	go func() {
		err := http.ListenAndServe(conf.FlagRunAddr, r)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			newLogger.Fatalf("server failed: %v", err)
		}
	}()

	resp, err := http.Get("http://localhost" + conf.FlagRunAddr)
	if err != nil {
		t.Fatalf("failed to get response from server: %v", err)
	}

	err = resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to close response body: %v", err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected status OK from server")

	cancel()
	time.Sleep(1 * time.Second)
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
