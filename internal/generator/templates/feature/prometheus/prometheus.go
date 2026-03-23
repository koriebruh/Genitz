package monitoring

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds application-level Prometheus counters and histograms.
type Metrics struct {
	Registry       *prometheus.Registry
	RequestTotal   *prometheus.CounterVec
	RequestLatency *prometheus.HistogramVec
}

// NewMetrics creates a new Prometheus registry with standard HTTP metrics pre-registered.
// Inject this into your handler/middleware layer.
func NewMetrics(namespace string) *Metrics {
	reg := prometheus.NewRegistry()

	requestTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests, partitioned by status and method.",
		},
		[]string{"method", "path", "status"},
	)

	requestLatency := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	reg.MustRegister(requestTotal, requestLatency)

	return &Metrics{
		Registry:       reg,
		RequestTotal:   requestTotal,
		RequestLatency: requestLatency,
	}
}

// Handler returns an HTTP handler that exposes metrics at /metrics.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.Registry, promhttp.HandlerOpts{})
}
