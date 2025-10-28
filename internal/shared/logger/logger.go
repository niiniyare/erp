package logger

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// LogLevel represents the severity of a log entry
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel // Default fallback
	}
}

// Fields represents structured logging fields
type Fields map[string]any

// Logger defines the interface for all logger implementations
type Logger interface {
	Debug(msg string, fields ...Fields)
	Info(msg string, fields ...Fields)
	Warn(msg string, fields ...Fields)
	Error(msg string, fields ...Fields)
	Fatal(msg string, fields ...Fields)

	DebugContext(ctx context.Context, msg string, fields ...Fields)
	InfoContext(ctx context.Context, msg string, fields ...Fields)
	WarnContext(ctx context.Context, msg string, fields ...Fields)
	ErrorContext(ctx context.Context, msg string, fields ...Fields)

	WithFields(fields Fields) Logger
	WithContext(ctx context.Context) Logger
	SetLevel(level LogLevel)
	Close() error
}

// LoggerType represents the type of logger to use
type LoggerType string

const (
	ZapLogger     LoggerType = "zap"
	ZerologLogger LoggerType = "zerolog"
	SlogLogger    LoggerType = "slog"
)

// Config holds configuration for the logger
type Config struct {
	Type        LoggerType
	Level       LogLevel
	Output      io.Writer
	Format      string // "json", "text", "console"
	Development bool
	ServiceName string
	Version     string
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	return Config{
		Type:        ZerologLogger,
		Level:       InfoLevel,
		Output:      os.Stdout,
		Format:      "console",
		Development: true,
		ServiceName: "app",
		Version:     "1.0.0",
	}
}

// LoggerFactory creates logger instances
type LoggerFactory struct{}

// NewLogger creates a new logger based on the configuration
func (f *LoggerFactory) NewLogger(config Config) (Logger, error) {
	switch config.Type {
	case ZapLogger:
		return newZapLogger(config)
	case ZerologLogger:
		return newZerologLogger(config)
	case SlogLogger:
		return newSlogLogger(config)
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", config.Type)
	}
}

// Global logger instance
var globalLogger Logger

// Initialize sets up the global logger
func Initialize(config Config) error {
	factory := &LoggerFactory{}
	logger, err := factory.NewLogger(config)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// InitializeFromEnv sets up the global logger from environment variables
func InitializeFromEnv() error {
	config := DefaultConfig()

	if loggerType := os.Getenv("LOG_TYPE"); loggerType != "" {
		config.Type = LoggerType(loggerType)
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		config.Level = ParseLogLevel(logLevel)
	}

	if logFormat := os.Getenv("LOG_FORMAT"); logFormat != "" {
		config.Format = logFormat
	}

	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		config.ServiceName = serviceName
	}

	if version := os.Getenv("SERVICE_VERSION"); version != "" {
		config.Version = version
	}

	if dev := os.Getenv("LOG_DEVELOPMENT"); dev == "true" {
		config.Development = true
	}

	return Initialize(config)
}

// Debug Global logging functions
func Debug(msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

func Info(msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

func Warn(msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

func Error(msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	}
}

func DebugContext(ctx context.Context, msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.DebugContext(ctx, msg, fields...)
	}
}

func InfoContext(ctx context.Context, msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.InfoContext(ctx, msg, fields...)
	}
}

func WarnContext(ctx context.Context, msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.WarnContext(ctx, msg, fields...)
	}
}

func ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	if globalLogger != nil {
		globalLogger.ErrorContext(ctx, msg, fields...)
	}
}

func WithFields(fields Fields) Logger {
	if globalLogger != nil {
		return globalLogger.WithFields(fields)
	}
	return nil
}

func WithContext(ctx context.Context) Logger {
	if globalLogger != nil {
		return globalLogger.WithContext(ctx)
	}
	return nil
}

func SetLevel(level LogLevel) {
	if globalLogger != nil {
		globalLogger.SetLevel(level)
	}
}

func Close() error {
	if globalLogger != nil {
		return globalLogger.Close()
	}
	return nil
}
