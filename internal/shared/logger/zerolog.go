package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

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
