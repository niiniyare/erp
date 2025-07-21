package metrics

import (
	"net/http"
	"time"
)

// MockMetricsService is a mock implementation of the MetricsProvider interface.
type MockMetricsService struct{}

// NewMockMetricsService creates and returns a new instance of MockMetricsService.
func NewMockMetricsService() *MockMetricsService {
	return &MockMetricsService{}
}

// Counter implements MetricsProvider.Counter.
func (m *MockMetricsService) Counter(name, help string, labelKeys ...string) Counter {
	return &noOpCounter{}
}

// IncrementCounter implements MetricsProvider.IncrementCounter.
func (m *MockMetricsService) IncrementCounter(name string, labels Fields) {
	// No-op for mock
}

// Gauge implements MetricsProvider.Gauge.
func (m *MockMetricsService) Gauge(name, help string, labelKeys ...string) Gauge {
	return &noOpGauge{}
}

// SetGauge implements MetricsProvider.SetGauge.
func (m *MockMetricsService) SetGauge(name string, value float64, labels Fields) {
	// No-op for mock
}

// Histogram implements MetricsProvider.Histogram.
func (m *MockMetricsService) Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram {
	return &noOpHistogram{}
}

// ObserveHistogram implements MetricsProvider.ObserveHistogram.
func (m *MockMetricsService) ObserveHistogram(name string, value float64, labels Fields) {
	// No-op for mock
}

// Timer implements MetricsProvider.Timer.
func (m *MockMetricsService) Timer(name string, labels Fields) Timer {
	return &noOpTimer{}
}

// TimerFunc implements MetricsProvider.TimerFunc.
func (m *MockMetricsService) TimerFunc(name string, labels Fields, fn func()) time.Duration {
	fn()
	return 0 // Return 0 duration for mock
}

// Handler implements MetricsProvider.Handler.
func (m *MockMetricsService) Handler() http.Handler {
	return http.NotFoundHandler()
}

// Close implements MetricsProvider.Close.
func (m *MockMetricsService) Close() error {
	return nil
}
