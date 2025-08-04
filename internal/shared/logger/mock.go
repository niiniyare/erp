package logger

import "context"

// MockLogger is a mock implementation of the Logger interface.
type MockLogger struct{}

// NewMockLogger creates and returns a new instance of MockLogger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Debug implements Logger.Debug.
func (m *MockLogger) Debug(msg string, fields ...Fields) {}

// Info implements Logger.Info.
func (m *MockLogger) Info(msg string, fields ...Fields) {}

// Warn implements Logger.Warn.
func (m *MockLogger) Warn(msg string, fields ...Fields) {}

// Error implements Logger.Error.
func (m *MockLogger) Error(msg string, fields ...Fields) {}

// Fatal implements Logger.Fatal.
func (m *MockLogger) Fatal(msg string, fields ...Fields) {}

// DebugContext implements Logger.DebugContext.
func (m *MockLogger) DebugContext(ctx context.Context, msg string, fields ...Fields) {}

// InfoContext implements Logger.InfoContext.
func (m *MockLogger) InfoContext(ctx context.Context, msg string, fields ...Fields) {}

// WarnContext implements Logger.WarnContext.
func (m *MockLogger) WarnContext(ctx context.Context, msg string, fields ...Fields) {}

// ErrorContext implements Logger.ErrorContext.
func (m *MockLogger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {}

// WithFields implements Logger.WithFields.
func (m *MockLogger) WithFields(fields Fields) Logger {
	return m
}

// WithContext implements Logger.WithContext.
func (m *MockLogger) WithContext(ctx context.Context) Logger {
	return m
}

// SetLevel implements Logger.SetLevel.
func (m *MockLogger) SetLevel(level LogLevel) {}

// Close implements Logger.Close.
func (m *MockLogger) Close() error {
	return nil
}
