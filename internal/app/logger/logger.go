package logger

import (
	"errors"

	"go.uber.org/zap"
)

func NewLogger(level string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()

	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, errors.New("logger parse level failed")
	}

	cfg.Level = atomicLevel

	zl, err := cfg.Build()

	if err != nil {
		return nil, errors.New("logger build failed")
	}

	sugar := zl.Sugar()

	return sugar, nil
}
