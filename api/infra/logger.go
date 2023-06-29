package infra

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes a zap logger with RFC3339 time format
func InitLogger() zap.Logger {
	var cfg zap.Config = zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return *logger
}
