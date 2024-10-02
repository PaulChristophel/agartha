package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger instance
var (
	logger *zap.Logger
	once   sync.Once
)

// InitLogger initializes and returns a `zap` logger based on the Gin mode
func InitLogger(mode string) (*zap.Logger, error) {
	var err error
	once.Do(func() {
		if mode == "debug" {
			logger, err = zap.NewDevelopment()
		} else {
			config := zap.NewProductionConfig()
			config.EncoderConfig.TimeKey = "timestamp"
			config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
			logger, err = config.Build()
		}
	})
	return logger, err
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if logger == nil {
		panic("Logger not initialized. Call InitLogger first.")
	}
	return logger
}
