// Package server exposes exporter metrics over HTTP.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/etnperlong/mihomo-exporter/internal/config"
)

var (
	errMissingGatherer     = errors.New("prometheus gatherer is required")
	errInvalidTelemetryURL = errors.New("web telemetry path must start with /")
)

// Server serves the exporter HTTP endpoints.
type Server struct {
	httpServer *http.Server
}

// New creates a Server using webConfig and gatherer.
func New(webConfig config.WebConfig, gatherer prometheus.Gatherer, logger *zap.Logger) (*Server, error) {
	if gatherer == nil {
		return nil, errMissingGatherer
	}
	if !strings.HasPrefix(webConfig.TelemetryPath, "/") {
		return nil, fmt.Errorf("validate telemetry path %q: %w", webConfig.TelemetryPath, errInvalidTelemetryURL)
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	mux := http.NewServeMux()
	mux.Handle(webConfig.TelemetryPath, promhttp.HandlerFor(gatherer, promhttp.HandlerOpts{
		ErrorLog: log.New(zapWriter{logger: logger.Named("promhttp")}, "", 0),
	}))
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)

			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(writer, "<a href=%q>Metrics</a>\n", webConfig.TelemetryPath)
	})

	return &Server{httpServer: &http.Server{
		Addr:              webConfig.ListenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Minute,
	}}, nil
}

// HTTPServer returns the configured underlying HTTP server.
func (s *Server) HTTPServer() *http.Server {
	return s.httpServer
}

// Serve accepts requests from listener until Shutdown closes the server.
func (s *Server) Serve(listener net.Listener) error {
	if err := s.httpServer.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP server: %w", err)
	}

	return nil
}

// ListenAndServe listens on the configured address until Shutdown closes the server.
func (s *Server) ListenAndServe() error {
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve HTTP server: %w", err)
	}

	return nil
}

// Shutdown gracefully stops accepting requests and waits for active requests.
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	return nil
}

// zapWriter routes promhttp's standard-library logger output to zap.
type zapWriter struct {
	logger *zap.Logger
}

func (w zapWriter) Write(message []byte) (int, error) {
	w.logger.Error("prometheus handler error", zap.String("message", strings.TrimSpace(string(message))))

	return len(message), nil
}
