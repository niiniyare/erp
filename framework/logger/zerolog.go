package logger

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

type zerologLogger struct {
	logger zerolog.Logger
	config Config
	level  LogLevel
}

func newZerologLogger(cfg Config) (*zerologLogger, error) {
	out := cfg.Output
	if cfg.Format == "console" || cfg.Format == "text" {
		out = zerolog.ConsoleWriter{Out: cfg.Output, TimeFormat: time.RFC3339}
	}

	l := zerolog.New(out).With().
		Timestamp().
		Str("app", cfg.ServiceName).
		Str("version", cfg.Version).
		Logger().
		Level(logLevelToZerolog(cfg.Level))

	return &zerologLogger{logger: l, config: cfg, level: cfg.Level}, nil
}

func (z *zerologLogger) Debug(msg string, fields ...Fields) {
	z.addFields(z.logger.Debug(), fields...).Msg(msg)
}
func (z *zerologLogger) Info(msg string, fields ...Fields) {
	z.addFields(z.logger.Info(), fields...).Msg(msg)
}
func (z *zerologLogger) Warn(msg string, fields ...Fields) {
	z.addFields(z.logger.Warn(), fields...).Msg(msg)
}
func (z *zerologLogger) Error(msg string, fields ...Fields) {
	z.addFields(z.logger.Error(), fields...).Msg(msg)
}

// Fatal logs at ERROR level with fatal=true. Does NOT call os.Exit.
func (z *zerologLogger) Fatal(msg string, fields ...Fields) {
	z.addFields(z.logger.Error().Bool("fatal", true), fields...).Msg(msg)
}

func (z *zerologLogger) DebugContext(_ context.Context, msg string, fields ...Fields) {
	z.Debug(msg, fields...)
}
func (z *zerologLogger) InfoContext(_ context.Context, msg string, fields ...Fields) {
	z.Info(msg, fields...)
}
func (z *zerologLogger) WarnContext(_ context.Context, msg string, fields ...Fields) {
	z.Warn(msg, fields...)
}
func (z *zerologLogger) ErrorContext(_ context.Context, msg string, fields ...Fields) {
	z.Error(msg, fields...)
}

func (z *zerologLogger) WithFields(fields Fields) Logger {
	ctx := z.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &zerologLogger{logger: ctx.Logger(), config: z.config, level: z.level}
}

func (z *zerologLogger) WithContext(_ context.Context) Logger {
	return &zerologLogger{logger: z.logger, config: z.config, level: z.level}
}

func (z *zerologLogger) SetLevel(level LogLevel) {
	z.level = level
	z.logger = z.logger.Level(logLevelToZerolog(level))
}

func (z *zerologLogger) Close() error { return nil }

func (z *zerologLogger) addFields(event *zerolog.Event, fields ...Fields) *zerolog.Event {
	for _, fm := range fields {
		for k, v := range fm {
			event = event.Interface(k, v)
		}
	}
	return event
}

func logLevelToZerolog(l LogLevel) zerolog.Level {
	switch l {
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
