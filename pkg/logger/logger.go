package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is the interface for logging operations
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync() error
}

// ZapLogger is the implementation of Logger using zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewLogger creates a new logger instance
func NewLogger(logLevel string, logFilePath string, maxSize, maxBackups, maxAge int, compress bool) (Logger, error) {
	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	// Configure file output with rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge,     // days
		Compress:   compress,
	})

	// Configure console output
	consoleWriter := zapcore.AddSync(os.Stdout)

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// Create cores
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		fileWriter,
		level,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		consoleWriter,
		level,
	)

	// Combine cores
	core := zapcore.NewTee(fileCore, consoleCore)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return &ZapLogger{logger: logger}, nil
}

// Info logs an info message
func (l *ZapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Error logs an error message
func (l *ZapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Warn logs a warning message
func (l *ZapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Debug logs a debug message
func (l *ZapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Fatal logs a fatal message and exits
func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Sync flushes any buffered log entries
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
