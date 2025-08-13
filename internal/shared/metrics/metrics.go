package metrics

//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Fields represents key-value pairs for labels/attributes
type Fields map[string]any

// MetricsProvider defines the interface for metrics collection
type MetricsProvider interface {
	// Counter operations
	Counter(name, help string, labelKeys ...string) Counter
	IncrementCounter(name string, labels Fields)

	// Gauge operations
	Gauge(name, help string, labelKeys ...string) Gauge
	SetGauge(name string, value float64, labels Fields)

	// Histogram operations
	Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram
	ObserveHistogram(name string, value float64, labels Fields)

	// Timer operations
	Timer(name string, labels Fields) Timer
	TimerFunc(name string, labels Fields, fn func()) time.Duration

	// HTTP handler for metrics exposure
	Handler() http.Handler

	// Close resources
	Close() error
}

// Counter interface
type Counter interface {
	Inc(labels Fields)
	Add(value float64, labels Fields)
}

// Gauge interface
type Gauge interface {
	Set(value float64, labels Fields)
	Inc(labels Fields)
	Dec(labels Fields)
	Add(value float64, labels Fields)
	Sub(value float64, labels Fields)
}

// Histogram interface
type Histogram interface {
	Observe(value float64, labels Fields)
}

// Timer interface
type Timer interface {
	Stop() time.Duration
}

// MetricsConfig holds configuration for metrics
type MetricsConfig struct {
	Provider  string // "prometheus" or "otel"
	Namespace string
	Subsystem string
	Enabled   bool
}

// MetricsService is the main service for managing metrics
type MetricsService struct {
	config   MetricsConfig
	provider MetricsProvider
	mu       sync.RWMutex
}

// NewMetricsService creates a new metrics service
func NewMetricsService(config MetricsConfig) (*MetricsService, error) {
	var provider MetricsProvider
	var err error

	if !config.Enabled {
		provider = &noOpProvider{}
	} else {
		switch config.Provider {
		case "prometheus":
			provider, err = NewPrometheusProvider(config.Namespace, config.Subsystem)
		case "otel":
			provider, err = NewOTelProvider(config.Namespace, config.Subsystem)
		default:
			return nil, fmt.Errorf("unsupported metrics provider: %s", config.Provider)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create metrics provider: %w", err)
	}

	return &MetricsService{
		config:   config,
		provider: provider,
	}, nil
}

// Counter creates or retrieves a counter metric
func (s *MetricsService) Counter(name, help string, labelKeys ...string) Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Counter(name, help, labelKeys...)
}

// IncrementCounter increments a counter by 1
func (s *MetricsService) IncrementCounter(name string, labels Fields) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.provider.IncrementCounter(name, labels)
}

// Gauge creates or retrieves a gauge metric
func (s *MetricsService) Gauge(name, help string, labelKeys ...string) Gauge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Gauge(name, help, labelKeys...)
}

// SetGauge sets a gauge value
func (s *MetricsService) SetGauge(name string, value float64, labels Fields) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.provider.SetGauge(name, value, labels)
}

// Histogram creates or retrieves a histogram metric
func (s *MetricsService) Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Histogram(name, help, buckets, labelKeys...)
}

// ObserveHistogram observes a value in a histogram
func (s *MetricsService) ObserveHistogram(name string, value float64, labels Fields) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.provider.ObserveHistogram(name, value, labels)
}

// Timer creates a timer for measuring durations
func (s *MetricsService) Timer(name string, labels Fields) Timer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Timer(name, labels)
}

// TimerFunc measures the duration of a function execution
func (s *MetricsService) TimerFunc(name string, labels Fields, fn func()) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.TimerFunc(name, labels, fn)
}

// Handler returns HTTP handler for metrics exposure
func (s *MetricsService) Handler() http.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.provider.Handler()
}

// Close closes the metrics service
func (s *MetricsService) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.provider.Close()
}

// Prometheus Implementation
type PrometheusProvider struct {
	registry   prometheus.Registerer
	gatherer   prometheus.Gatherer
	namespace  string
	subsystem  string
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	mu         sync.RWMutex
}

func NewPrometheusProvider(namespace, subsystem string) (*PrometheusProvider, error) {
	registry := prometheus.NewRegistry()

	return &PrometheusProvider{
		registry:   registry,
		gatherer:   registry,
		namespace:  namespace,
		subsystem:  subsystem,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}, nil
}

func (p *PrometheusProvider) Counter(name, help string, labelKeys ...string) Counter {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if counter, exists := p.counters[key]; exists {
		return &prometheusCounter{counter: counter}
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
		},
		labelKeys,
	)

	if err := p.registry.Register(counter); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(*prometheus.CounterVec); ok {
				p.counters[key] = existing
				return &prometheusCounter{counter: existing}
			}
		}
		return &noOpCounter{}
	}

	p.counters[key] = counter
	return &prometheusCounter{counter: counter}
}

func (p *PrometheusProvider) IncrementCounter(name string, labels Fields) {
	labelKeys := extractLabelKeys(labels)
	counter := p.Counter(name, fmt.Sprintf("Auto-generated counter for %s", name), labelKeys...)
	counter.Inc(labels)
}

func (p *PrometheusProvider) Gauge(name, help string, labelKeys ...string) Gauge {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if gauge, exists := p.gauges[key]; exists {
		return &prometheusGauge{gauge: gauge}
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
		},
		labelKeys,
	)

	if err := p.registry.Register(gauge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(*prometheus.GaugeVec); ok {
				p.gauges[key] = existing
				return &prometheusGauge{gauge: existing}
			}
		}
		return &noOpGauge{}
	}

	p.gauges[key] = gauge
	return &prometheusGauge{gauge: gauge}
}

func (p *PrometheusProvider) SetGauge(name string, value float64, labels Fields) {
	labelKeys := extractLabelKeys(labels)
	gauge := p.Gauge(name, fmt.Sprintf("Auto-generated gauge for %s", name), labelKeys...)
	gauge.Set(value, labels)
}

func (p *PrometheusProvider) Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := p.metricKey(name)
	if histogram, exists := p.histograms[key]; exists {
		return &prometheusHistogram{histogram: histogram}
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labelKeys,
	)

	if err := p.registry.Register(histogram); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			if existing, ok := are.ExistingCollector.(*prometheus.HistogramVec); ok {
				p.histograms[key] = existing
				return &prometheusHistogram{histogram: existing}
			}
		}
		return &noOpHistogram{}
	}

	p.histograms[key] = histogram
	return &prometheusHistogram{histogram: histogram}
}

func (p *PrometheusProvider) ObserveHistogram(name string, value float64, labels Fields) {
	labelKeys := extractLabelKeys(labels)
	histogram := p.Histogram(name, fmt.Sprintf("Auto-generated histogram for %s", name), nil, labelKeys...)
	histogram.Observe(value, labels)
}

func (p *PrometheusProvider) Timer(name string, labels Fields) Timer {
	labelKeys := extractLabelKeys(labels)
	histogram := p.Histogram(name+"_duration_seconds", fmt.Sprintf("Duration histogram for %s", name), nil, labelKeys...)
	return &prometheusTimer{
		histogram: histogram,
		labels:    labels,
		start:     time.Now(),
	}
}

func (p *PrometheusProvider) TimerFunc(name string, labels Fields, fn func()) time.Duration {
	timer := p.Timer(name, labels)
	fn()
	return timer.Stop()
}

func (p *PrometheusProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.gatherer, promhttp.HandlerOpts{})
}

func (p *PrometheusProvider) Close() error {
	return nil
}

func (p *PrometheusProvider) metricKey(name string) string {
	if p.subsystem != "" {
		return fmt.Sprintf("%s_%s_%s", p.namespace, p.subsystem, name)
	}
	return fmt.Sprintf("%s_%s", p.namespace, name)
}

// OpenTelemetry Implementation
type OTelProvider struct {
	meter      metric.Meter
	counters   map[string]metric.Int64Counter
	gauges     map[string]metric.Float64Gauge
	histograms map[string]metric.Float64Histogram
	mu         sync.RWMutex
}

func NewOTelProvider(namespace, subsystem string) (*OTelProvider, error) {
	// Note: In a real implementation, you would get the meter from your OTel setup
	// meter := otel.Meter(namespace)

	return &OTelProvider{
		// meter:      meter,
		counters:   make(map[string]metric.Int64Counter),
		gauges:     make(map[string]metric.Float64Gauge),
		histograms: make(map[string]metric.Float64Histogram),
	}, nil
}

func (o *OTelProvider) Counter(name, help string, labelKeys ...string) Counter {
	o.mu.Lock()
	defer o.mu.Unlock()

	if counter, exists := o.counters[name]; exists {
		return &otelCounter{counter: counter}
	}

	counter, err := o.meter.Int64Counter(name, metric.WithDescription(help))
	if err != nil {
		return &noOpCounter{}
	}

	o.counters[name] = counter
	return &otelCounter{counter: counter}
}

func (o *OTelProvider) IncrementCounter(name string, labels Fields) {
	counter := o.Counter(name, fmt.Sprintf("Auto-generated counter for %s", name))
	counter.Inc(labels)
}

func (o *OTelProvider) Gauge(name, help string, labelKeys ...string) Gauge {
	o.mu.Lock()
	defer o.mu.Unlock()

	if gauge, exists := o.gauges[name]; exists {
		return &otelGauge{gauge: gauge}
	}

	gauge, err := o.meter.Float64Gauge(name, metric.WithDescription(help))
	if err != nil {
		return &noOpGauge{}
	}

	o.gauges[name] = gauge
	return &otelGauge{gauge: gauge}
}

func (o *OTelProvider) SetGauge(name string, value float64, labels Fields) {
	gauge := o.Gauge(name, fmt.Sprintf("Auto-generated gauge for %s", name))
	gauge.Set(value, labels)
}

func (o *OTelProvider) Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram {
	o.mu.Lock()
	defer o.mu.Unlock()

	if histogram, exists := o.histograms[name]; exists {
		return &otelHistogram{histogram: histogram}
	}

	histogram, err := o.meter.Float64Histogram(name, metric.WithDescription(help))
	if err != nil {
		return &noOpHistogram{}
	}

	o.histograms[name] = histogram
	return &otelHistogram{histogram: histogram}
}

func (o *OTelProvider) ObserveHistogram(name string, value float64, labels Fields) {
	histogram := o.Histogram(name, fmt.Sprintf("Auto-generated histogram for %s", name), nil)
	histogram.Observe(value, labels)
}

func (o *OTelProvider) Timer(name string, labels Fields) Timer {
	histogram := o.Histogram(name+"_duration", fmt.Sprintf("Duration histogram for %s", name), nil)
	return &otelTimer{
		histogram: histogram,
		labels:    labels,
		start:     time.Now(),
	}
}

func (o *OTelProvider) TimerFunc(name string, labels Fields, fn func()) time.Duration {
	timer := o.Timer(name, labels)
	fn()
	return timer.Stop()
}

func (o *OTelProvider) Handler() http.Handler {
	return http.NotFoundHandler()
}

func (o *OTelProvider) Close() error {
	return nil
}

// Prometheus metric implementations
type prometheusCounter struct {
	counter *prometheus.CounterVec
}

func (c *prometheusCounter) Inc(labels Fields) {
	c.counter.With(fieldsToPrometheusLabels(labels)).Inc()
}

func (c *prometheusCounter) Add(value float64, labels Fields) {
	c.counter.With(fieldsToPrometheusLabels(labels)).Add(value)
}

type prometheusGauge struct {
	gauge *prometheus.GaugeVec
}

func (g *prometheusGauge) Set(value float64, labels Fields) {
	g.gauge.With(fieldsToPrometheusLabels(labels)).Set(value)
}

func (g *prometheusGauge) Inc(labels Fields) {
	g.gauge.With(fieldsToPrometheusLabels(labels)).Inc()
}

func (g *prometheusGauge) Dec(labels Fields) {
	g.gauge.With(fieldsToPrometheusLabels(labels)).Dec()
}

func (g *prometheusGauge) Add(value float64, labels Fields) {
	g.gauge.With(fieldsToPrometheusLabels(labels)).Add(value)
}

func (g *prometheusGauge) Sub(value float64, labels Fields) {
	g.gauge.With(fieldsToPrometheusLabels(labels)).Sub(value)
}

type prometheusHistogram struct {
	histogram *prometheus.HistogramVec
}

func (h *prometheusHistogram) Observe(value float64, labels Fields) {
	h.histogram.With(fieldsToPrometheusLabels(labels)).Observe(value)
}

type prometheusTimer struct {
	histogram Histogram
	labels    Fields
	start     time.Time
}

func (t *prometheusTimer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.histogram.Observe(duration.Seconds(), t.labels)
	return duration
}

// OpenTelemetry metric implementations
type otelCounter struct {
	counter metric.Int64Counter
}

func (c *otelCounter) Inc(labels Fields) {
	c.counter.Add(context.Background(), 1, metric.WithAttributes(fieldsToAttributes(labels)...))
}

func (c *otelCounter) Add(value float64, labels Fields) {
	c.counter.Add(context.Background(), int64(value), metric.WithAttributes(fieldsToAttributes(labels)...))
}

type otelGauge struct {
	gauge metric.Float64Gauge
}

func (g *otelGauge) Set(value float64, labels Fields) {
	g.gauge.Record(context.Background(), value, metric.WithAttributes(fieldsToAttributes(labels)...))
}

func (g *otelGauge) Inc(labels Fields) {
	g.Set(1, labels)
}

func (g *otelGauge) Dec(labels Fields) {
	g.Set(-1, labels)
}

func (g *otelGauge) Add(value float64, labels Fields) {
	g.Set(value, labels)
}

func (g *otelGauge) Sub(value float64, labels Fields) {
	g.Set(-value, labels)
}

type otelHistogram struct {
	histogram metric.Float64Histogram
}

func (h *otelHistogram) Observe(value float64, labels Fields) {
	h.histogram.Record(context.Background(), value, metric.WithAttributes(fieldsToAttributes(labels)...))
}

type otelTimer struct {
	histogram Histogram
	labels    Fields
	start     time.Time
}

func (t *otelTimer) Stop() time.Duration {
	duration := time.Since(t.start)
	t.histogram.Observe(duration.Seconds(), t.labels)
	return duration
}

// No-op implementations
type noOpProvider struct{}

func (n *noOpProvider) Counter(name, help string, labelKeys ...string) Counter { return &noOpCounter{} }
func (n *noOpProvider) IncrementCounter(name string, labels Fields)            {}
func (n *noOpProvider) Gauge(name, help string, labelKeys ...string) Gauge     { return &noOpGauge{} }
func (n *noOpProvider) SetGauge(name string, value float64, labels Fields)     {}
func (n *noOpProvider) Histogram(name, help string, buckets []float64, labelKeys ...string) Histogram {
	return &noOpHistogram{}
}
func (n *noOpProvider) ObserveHistogram(name string, value float64, labels Fields) {}
func (n *noOpProvider) Timer(name string, labels Fields) Timer                     { return &noOpTimer{} }
func (n *noOpProvider) TimerFunc(name string, labels Fields, fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}
func (n *noOpProvider) Handler() http.Handler { return http.NotFoundHandler() }
func (n *noOpProvider) Close() error          { return nil }

type noOpCounter struct{}

func (n *noOpCounter) Inc(labels Fields)                {}
func (n *noOpCounter) Add(value float64, labels Fields) {}

type noOpGauge struct{}

func (n *noOpGauge) Set(value float64, labels Fields) {}
func (n *noOpGauge) Inc(labels Fields)                {}
func (n *noOpGauge) Dec(labels Fields)                {}
func (n *noOpGauge) Add(value float64, labels Fields) {}
func (n *noOpGauge) Sub(value float64, labels Fields) {}

type noOpHistogram struct{}

func (n *noOpHistogram) Observe(value float64, labels Fields) {}

type noOpTimer struct{}

func (n *noOpTimer) Stop() time.Duration { return 0 }

// Helper functions
func fieldsToPrometheusLabels(fields Fields) prometheus.Labels {
	labels := make(prometheus.Labels, len(fields))
	for k, v := range fields {
		labels[k] = fmt.Sprintf("%v", v)
	}
	return labels
}

func fieldsToAttributes(fields Fields) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(fields))
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", val)))
		}
	}
	return attrs
}

func extractLabelKeys(labels Fields) []string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	return keys
}

// Predefined bucket configurations
func StandardHTTPDurationBuckets() []float64 {
	return []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
}

func StandardSizeBuckets() []float64 {
	return []float64{100, 1000, 10000, 100000, 1000000, 10000000, 100000000}
}
