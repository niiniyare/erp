package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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
	zapConfig.InitialFields = map[string]any{
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
