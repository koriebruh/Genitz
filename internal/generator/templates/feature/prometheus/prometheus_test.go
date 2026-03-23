package monitoring

import (
	"net/http/httptest"
	"testing"
)

func TestNewMetrics_CreatesRegistry(t *testing.T) {
	m := NewMetrics("myapp")
	if m == nil {
		t.Fatal("Expected non-nil *Metrics")
	}
	if m.Registry == nil {
		t.Error("Expected non-nil prometheus.Registry")
	}
	if m.RequestTotal == nil {
		t.Error("Expected non-nil RequestTotal counter")
	}
	if m.RequestLatency == nil {
		t.Error("Expected non-nil RequestLatency histogram")
	}
}

func TestMetrics_RecordRequest(t *testing.T) {
	m := NewMetrics("test")

	// Record a request
	m.RequestTotal.WithLabelValues("GET", "/health", "200").Inc()

	// Verify the handler exposes metrics
	handler := m.Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200 from /metrics, got %d", w.Code)
	}
	body := w.Body.String()
	if len(body) == 0 {
		t.Error("Expected non-empty metrics output")
	}
}
