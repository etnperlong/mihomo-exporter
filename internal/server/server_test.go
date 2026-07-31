package server_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/etnperlong/mihomo-exporter/internal/config"
	"github.com/etnperlong/mihomo-exporter/internal/server"
)

func TestNew(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	metric := prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_metric", Help: "test metric"})
	registry.MustRegister(metric)
	metric.Set(1)

	srv, err := server.New(config.WebConfig{ListenAddress: "127.0.0.1:0", TelemetryPath: "/metrics"}, registry, nil)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	response := httptest.NewRecorder()
	srv.HTTPServer().Handler.ServeHTTP(response, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("GET /metrics status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "test_metric 1") {
		t.Errorf("GET /metrics body = %q, want test metric", response.Body.String())
	}
}

func TestNewTimeouts(t *testing.T) {
	t.Parallel()

	srv, err := server.New(config.WebConfig{TelemetryPath: "/metrics"}, prometheus.NewRegistry(), nil)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	got := srv.HTTPServer()
	if got, want := got.ReadHeaderTimeout, 5*time.Second; got != want {
		t.Errorf("ReadHeaderTimeout = %s, want %s", got, want)
	}
	if got, want := got.ReadTimeout, 15*time.Second; got != want {
		t.Errorf("ReadTimeout = %s, want %s", got, want)
	}
	if got, want := got.WriteTimeout, 30*time.Second; got != want {
		t.Errorf("WriteTimeout = %s, want %s", got, want)
	}
	if got, want := got.IdleTimeout, time.Minute; got != want {
		t.Errorf("IdleTimeout = %s, want %s", got, want)
	}
}

func TestServer_Shutdown(t *testing.T) {
	t.Parallel()

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() returned unexpected error: %v", err)
	}
	srv, err := server.New(config.WebConfig{TelemetryPath: "/metrics"}, prometheus.NewRegistry(), nil)
	if err != nil {
		t.Fatalf("New() returned unexpected error: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- srv.Serve(listener)
	}()
	if err := srv.Shutdown(t.Context()); err != nil {
		t.Fatalf("Shutdown() returned unexpected error: %v", err)
	}
	if err := <-done; err != nil {
		t.Errorf("Serve() returned unexpected error: %v", err)
	}
}
