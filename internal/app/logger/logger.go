package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(level zapcore.Level) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	cfg.Level.SetLevel(level)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger build failed: %w", err)
	}

	defer func() {
		err := logger.Sync()

		if err != nil {
			logger.Error("failed to sync logger", zap.Error(err))
		}
	}()

	return logger, nil
}
