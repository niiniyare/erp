// Package logger provides a structured logging abstraction for the Awo Framework.
// Supported backends: zerolog (default), zap, slog.
package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
)

// LogLevel represents the severity of a log entry.
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
		return InfoLevel
	}
}

// Fields represents structured logging fields.
type Fields map[string]any

// Logger is the framework logger interface.
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

// LoggerType selects the backend implementation.
type LoggerType string

const (
	ZerologLogger LoggerType = "zerolog"
	ZapLogger     LoggerType = "zap"
	SlogLogger    LoggerType = "slog"
)

// Config holds logger configuration.
type Config struct {
	Type        LoggerType
	Level       LogLevel
	Output      io.Writer
	Format      string // "json", "text", "console"
	Development bool
	ServiceName string
	Version     string
}

func DefaultConfig() Config {
	return Config{
		Type:        ZerologLogger,
		Level:       InfoLevel,
		Output:      os.Stdout,
		Format:      "json",
		ServiceName: "app",
		Version:     "1.0.0",
	}
}

// New creates a Logger from config.
func New(cfg Config) (Logger, error) {
	switch cfg.Type {
	case ZerologLogger:
		return newZerologLogger(cfg)
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", cfg.Type)
	}
}

// NewNoOp returns a Logger that discards all output. Useful for tests.
func NewNoOp() Logger { return noopLogger{} }

type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...Fields)                           {}
func (noopLogger) Info(_ string, _ ...Fields)                            {}
func (noopLogger) Warn(_ string, _ ...Fields)                            {}
func (noopLogger) Error(_ string, _ ...Fields)                           {}
func (noopLogger) Fatal(_ string, _ ...Fields)                           {}
func (noopLogger) DebugContext(_ context.Context, _ string, _ ...Fields) {}
func (noopLogger) InfoContext(_ context.Context, _ string, _ ...Fields)  {}
func (noopLogger) WarnContext(_ context.Context, _ string, _ ...Fields)  {}
func (noopLogger) ErrorContext(_ context.Context, _ string, _ ...Fields) {}
func (noopLogger) WithFields(_ Fields) Logger                            { return noopLogger{} }
func (noopLogger) WithContext(_ context.Context) Logger                  { return noopLogger{} }
func (noopLogger) SetLevel(_ LogLevel)                                   {}
func (noopLogger) Close() error                                          { return nil }
