package logging

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger is the global Zap logger instance
var Logger *zap.Logger

// Init initializes the Zap logger with production settings
func Init() {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.Level = zap.NewAtomicLevelAt(zap.InfoLevel) // Configurable later

    logger, err := config.Build()
    if err != nil {
        panic("Failed to initialize logger: " + err.Error())
    }
    Logger = logger
}

// Sync flushes any buffered log entries
func Sync() {
    if Logger != nil {
        Logger.Sync()
    }
}