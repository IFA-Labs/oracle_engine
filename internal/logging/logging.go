package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the global Zap logger instance
var Logger *zap.Logger

// Init initializes the Zap logger with production settings
func Init() {
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder

	encoder := zapcore.NewJSONEncoder(config)

	errorSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/error.log",
		MaxSize:    20,
		MaxBackups: 7,
		MaxAge:     45,   // days
		Compress:   true, // disabled by default
	})
	errorCore := zapcore.NewCore(encoder, errorSyncer, zapcore.ErrorLevel)

	infoSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/info.log",
		MaxSize:    10,
		MaxBackups: 4,
		MaxAge:     30,   // days
		Compress:   true, // disabled by default
	})
	infoCore := zapcore.NewCore(encoder, infoSyncer, zapcore.InfoLevel)

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)

	logger := zap.New(zapcore.NewTee(
		errorCore,
		infoCore,
		consoleCore,
	))
	Logger = logger
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

