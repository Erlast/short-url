package main

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/Erlast/short-url.git/internal/app/logger"
	"github.com/Erlast/short-url.git/internal/app/storages"
)

func InitTestCfg(t *testing.T) (*config.Cfg, storages.URLStorage, *zap.SugaredLogger) {
	t.Helper()

	conf := &config.Cfg{
		FlagRunAddr: ":8080",
		FlagBaseURL: "http://localhost:8080",
	}
	ctx := context.Background()

	ctx = context.WithValue(ctx, helpers.UserID, uuid.NewString())

	newLogger, err := logger.NewLogger("info")

	if err != nil {
		t.Errorf("failed to initialize test cfg (logger): %v", err)
		return nil, nil, nil
	}
	store, err := storages.NewStorage(ctx, conf, newLogger)

	if err != nil {
		t.Errorf("failed to initialize test cfg (storage): %v", err)
		return nil, nil, nil
	}

	return conf, store, newLogger
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
