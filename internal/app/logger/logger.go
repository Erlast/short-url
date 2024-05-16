package logger

import (
	"errors"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func NewLogger(level string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()

	cfg.Level, _ = zap.ParseAtomicLevel(level)

	zl, err := cfg.Build()

	if err != nil {
		return nil, errors.New("logger build failed")
	}

	sugar := zl.Sugar()

	Log = sugar

	return sugar, nil
}
