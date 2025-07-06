package logger

import (
	"context"
	"log/slog"
	"os"
)

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
