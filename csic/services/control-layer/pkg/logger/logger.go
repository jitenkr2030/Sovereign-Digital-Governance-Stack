package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log     *zap.Logger
	logOnce sync.Once
)

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Sync()
}

// ZapLogger wraps zap.Logger to implement Logger interface
type ZapLogger struct {
	*zap.Logger
}

// NewLogger creates a basic logger for early startup
func NewLogger() *zap.Logger {
	logOnce.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

		var err error
		log, err = config.Build()
		if err != nil {
			panic("failed to initialize logger: " + err.Error())
		}
	})

	return log
}

// NewZapLogger creates a configurable logger
func NewZapLogger(level string) (*zap.Logger, error) {
	var zapLevel zapcore.Level

	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	return config.Build()
}

// String creates a string field
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates an int field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64 creates a float64 field
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Bool creates a bool field
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Error creates an error field
func Error(err error) zap.Field {
	return zap.Error(err)
}

// Any creates a field for any type
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// FatalOnError logs fatal if error is not nil
func FatalOnError(logger *zap.Logger, err error, msg string) {
	if err != nil {
		logger.Fatal(msg, zap.Error(err))
	}
}

// Sync flushes any buffered log entries
func Sync(logger *zap.Logger) {
	if logger != nil {
		logger.Sync()
	}
}

// GetLevelFromString converts string level to zapcore.Level
func GetLevelFromString(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// WithEnvSuffix adds environment suffix to logger
func WithEnvSuffix(logger *zap.Logger, env string) *zap.Logger {
	if env == "production" {
		return logger
	}
	return logger.With(zap.String("env", env))
}

// NewNopLogger returns a no-op logger for testing
func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}

// FatalError checks error and calls fatal if not nil
func FatalError(logger *zap.Logger, err error, msg string) {
	if err != nil {
		logger.Fatal(msg, zap.Error(err))
		os.Exit(1)
	}
}
