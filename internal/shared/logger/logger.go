package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// Fields represents structured logging fields
type Fields map[string]interface{}

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
		Type:        SlogLogger,
		Level:       InfoLevel,
		Output:      os.Stdout,
		Format:      "json",
		Development: false,
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

// Zap Logger Implementation
type zapLogger struct {
	logger *zap.Logger
	config Config
}

func newZapLogger(config Config) (*zapLogger, error) {
	var zapConfig zap.Config

	if config.Development {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	// Set format
	if config.Format == "console" || config.Format == "text" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	// Set level
	zapConfig.Level = zap.NewAtomicLevelAt(logLevelToZap(config.Level))

	// Add initial fields
	zapConfig.InitialFields = map[string]interface{}{
		"service": config.ServiceName,
		"version": config.Version,
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		logger: logger,
		config: config,
	}, nil
}

func (z *zapLogger) Debug(msg string, fields ...Fields) {
	z.logger.Debug(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) Info(msg string, fields ...Fields) {
	z.logger.Info(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) Warn(msg string, fields ...Fields) {
	z.logger.Warn(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) Error(msg string, fields ...Fields) {
	z.logger.Error(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) Fatal(msg string, fields ...Fields) {
	z.logger.Fatal(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) DebugContext(ctx context.Context, msg string, fields ...Fields) {
	z.logger.Debug(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) InfoContext(ctx context.Context, msg string, fields ...Fields) {
	z.logger.Info(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) WarnContext(ctx context.Context, msg string, fields ...Fields) {
	z.logger.Warn(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	z.logger.Error(msg, z.fieldsToZap(fields...)...)
}

func (z *zapLogger) WithFields(fields Fields) Logger {
	return &zapLogger{
		logger: z.logger.With(z.fieldsToZap(fields)...),
		config: z.config,
	}
}

func (z *zapLogger) WithContext(ctx context.Context) Logger {
	return z // Zap doesn't have built-in context support
}

func (z *zapLogger) SetLevel(level LogLevel) {
	// Note: This changes the global level, not per-instance
	z.logger.Core().Enabled(logLevelToZap(level))
}

func (z *zapLogger) Close() error {
	return z.logger.Sync()
}

func (z *zapLogger) fieldsToZap(fields ...Fields) []zap.Field {
	var zapFields []zap.Field
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			zapFields = append(zapFields, zap.Any(k, v))
		}
	}
	return zapFields
}

func logLevelToZap(level LogLevel) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Zerolog Logger Implementation
type zerologLogger struct {
	logger zerolog.Logger
	config Config
	level  LogLevel
}

func newZerologLogger(config Config) (*zerologLogger, error) {
	output := config.Output
	if config.Format == "console" || config.Format == "text" {
		output = zerolog.ConsoleWriter{Out: config.Output, TimeFormat: time.RFC3339}
	}

	logger := zerolog.New(output).With().
		Timestamp().
		Str("service", config.ServiceName).
		Str("version", config.Version).
		Logger()

	// Set level
	logger = logger.Level(logLevelToZerolog(config.Level))

	return &zerologLogger{
		logger: logger,
		config: config,
		level:  config.Level,
	}, nil
}

func (z *zerologLogger) Debug(msg string, fields ...Fields) {
	event := z.logger.Debug()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) Info(msg string, fields ...Fields) {
	event := z.logger.Info()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) Warn(msg string, fields ...Fields) {
	event := z.logger.Warn()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) Error(msg string, fields ...Fields) {
	event := z.logger.Error()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) Fatal(msg string, fields ...Fields) {
	event := z.logger.Fatal()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) DebugContext(ctx context.Context, msg string, fields ...Fields) {
	event := z.logger.Debug()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) InfoContext(ctx context.Context, msg string, fields ...Fields) {
	event := z.logger.Info()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) WarnContext(ctx context.Context, msg string, fields ...Fields) {
	event := z.logger.Warn()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	event := z.logger.Error()
	z.addFields(event, fields...)
	event.Msg(msg)
}

func (z *zerologLogger) WithFields(fields Fields) Logger {
	ctx := z.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &zerologLogger{
		logger: ctx.Logger(),
		config: z.config,
		level:  z.level,
	}
}

func (z *zerologLogger) WithContext(ctx context.Context) Logger {
	return &zerologLogger{
		logger: z.logger.With().Logger(),
		config: z.config,
		level:  z.level,
	}
}

func (z *zerologLogger) SetLevel(level LogLevel) {
	z.level = level
	z.logger = z.logger.Level(logLevelToZerolog(level))
}

func (z *zerologLogger) Close() error {
	return nil // Zerolog doesn't require explicit closing
}

func (z *zerologLogger) addFields(event *zerolog.Event, fields ...Fields) {
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			event.Interface(k, v)
		}
	}
}

func logLevelToZerolog(level LogLevel) zerolog.Level {
	switch level {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Slog Logger Implementation
type slogLogger struct {
	logger *slog.Logger
	config Config
	level  LogLevel
}

func newSlogLogger(config Config) (*slogLogger, error) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: logLevelToSlog(config.Level),
	}

	if config.Format == "console" || config.Format == "text" {
		handler = slog.NewTextHandler(config.Output, opts)
	} else {
		handler = slog.NewJSONHandler(config.Output, opts)
	}

	logger := slog.New(handler).With(
		"service", config.ServiceName,
		"version", config.Version,
	)

	return &slogLogger{
		logger: logger,
		config: config,
		level:  config.Level,
	}, nil
}

func (s *slogLogger) Debug(msg string, fields ...Fields) {
	s.logger.Debug(msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) Info(msg string, fields ...Fields) {
	s.logger.Info(msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) Warn(msg string, fields ...Fields) {
	s.logger.Warn(msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) Error(msg string, fields ...Fields) {
	s.logger.Error(msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) Fatal(msg string, fields ...Fields) {
	s.logger.Error(msg, s.fieldsToSlog(fields...)...) // slog doesn't have Fatal
	os.Exit(1)
}

func (s *slogLogger) DebugContext(ctx context.Context, msg string, fields ...Fields) {
	s.logger.DebugContext(ctx, msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) InfoContext(ctx context.Context, msg string, fields ...Fields) {
	s.logger.InfoContext(ctx, msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) WarnContext(ctx context.Context, msg string, fields ...Fields) {
	s.logger.WarnContext(ctx, msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	s.logger.ErrorContext(ctx, msg, s.fieldsToSlog(fields...)...)
}

func (s *slogLogger) WithFields(fields Fields) Logger {
	return &slogLogger{
		logger: s.logger.With(s.fieldsToSlog(fields)...),
		config: s.config,
		level:  s.level,
	}
}

func (s *slogLogger) WithContext(ctx context.Context) Logger {
	return s // Return same instance as slog methods accept context
}

func (s *slogLogger) SetLevel(level LogLevel) {
	s.level = level
	// Note: slog level is set at handler creation, can't be changed dynamically
}

func (s *slogLogger) Close() error {
	return nil // slog doesn't require explicit closing
}

func (s *slogLogger) fieldsToSlog(fields ...Fields) []any {
	var slogArgs []any
	for _, fieldMap := range fields {
		for k, v := range fieldMap {
			slogArgs = append(slogArgs, k, v)
		}
	}
	return slogArgs
}

func logLevelToSlog(level LogLevel) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case FatalLevel:
		return slog.LevelError // slog doesn't have Fatal level
	default:
		return slog.LevelInfo
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
		switch strings.ToLower(logLevel) {
		case "debug":
			config.Level = DebugLevel
		case "info":
			config.Level = InfoLevel
		case "warn", "warning":
			config.Level = WarnLevel
		case "error":
			config.Level = ErrorLevel
		case "fatal":
			config.Level = FatalLevel
		}
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

// Global logging functions
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
