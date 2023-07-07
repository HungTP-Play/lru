package shared

import (
	"bytes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

type Metrics struct{}

func NewMetrics() *Metrics {
	return &Metrics{}
}

// RegisterCounter register counter
func (m *Metrics) RegisterCounter(name string, help string, labels []string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)

	prometheus.MustRegister(counter)
	return counter
}

// RegisterGauge register gauge
func (m *Metrics) RegisterGauge(name string, help string, labels []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: help,
	}, labels)

	prometheus.MustRegister(gauge)
	return gauge
}

// RegisterHistogram register histogram
func (m *Metrics) RegisterHistogram(name string, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)

	prometheus.MustRegister(histogram)
	return histogram
}

// RegisterSummary register summary
func (m *Metrics) RegisterSummary(name string, help string, labels []string, objectives map[float64]float64) *prometheus.SummaryVec {
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       name,
		Help:       help,
		Objectives: objectives,
	}, labels)

	prometheus.MustRegister(summary)
	return summary
}

// IncCounter increase counter
func (m *Metrics) IncCounter(counter *prometheus.CounterVec, labels ...string) {
	counter.WithLabelValues(labels...).Inc()
}

// IncGauge increase gauge
func (m *Metrics) IncGauge(gauge *prometheus.GaugeVec, labels ...string) {
	gauge.WithLabelValues(labels...).Inc()
}

// DecGauge decrease gauge
func (m *Metrics) DecGauge(gauge *prometheus.GaugeVec, labels ...string) {
	gauge.WithLabelValues(labels...).Dec()
}

// SetGauge set gauge
func (m *Metrics) SetGauge(gauge *prometheus.GaugeVec, value float64, labels ...string) {
	gauge.WithLabelValues(labels...).Set(value)
}

// ObserveHistogram observe histogram
func (m *Metrics) ObserveHistogram(histogram *prometheus.HistogramVec, value float64, labels ...string) {
	histogram.WithLabelValues(labels...).Observe(value)
}

// ObserveSummary observe summary
func (m *Metrics) ObserveSummary(summary *prometheus.SummaryVec, value float64, labels ...string) {
	summary.WithLabelValues(labels...).Observe(value)
}

// GetCounter get counter
func (m *Metrics) GetCounter(counter *prometheus.CounterVec, labels ...string) prometheus.Counter {
	return counter.WithLabelValues(labels...)
}

// GetGauge get gauge
func (m *Metrics) GetGauge(gauge *prometheus.GaugeVec, labels ...string) prometheus.Gauge {
	return gauge.WithLabelValues(labels...)
}

// GetHistogram get histogram
func (m *Metrics) GetHistogram(histogram *prometheus.HistogramVec, labels ...string) prometheus.Observer {
	return histogram.WithLabelValues(labels...)
}

// GetSummary get summary
func (m *Metrics) GetSummary(summary *prometheus.SummaryVec, labels ...string) prometheus.Observer {
	return summary.WithLabelValues(labels...)
}

func (m *Metrics) GetPrometheusMetrics() (string, error) {
	var metrics bytes.Buffer
	reg := prometheus.DefaultRegisterer.(*prometheus.Registry)
	metricFamilies, err := reg.Gather()
	if err != nil {
		return "", err
	}

	encoder := expfmt.NewEncoder(&metrics, expfmt.FmtText)
	for _, mf := range metricFamilies {
		if err := encoder.Encode(mf); err != nil {
			return "", err
		}
	}

	return metrics.String(), nil
}
