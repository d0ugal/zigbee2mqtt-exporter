package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server
type Server struct {
	cfg     *config.Config
	metrics *metrics.Registry
	server  *http.Server
}

// New creates a new server instance
func New(cfg *config.Config, metrics *metrics.Registry) *Server {
	s := &Server{
		cfg:     cfg,
		metrics: metrics,
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		versionInfo := version.Get()

		response := map[string]interface{}{
			"status":     "healthy",
			"timestamp":  time.Now().Unix(),
			"service":    "zigbee2mqtt-exporter",
			"version":    versionInfo.Version,
			"commit":     versionInfo.Commit,
			"build_date": versionInfo.BuildDate,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		jsonData, err := json.Marshal(response)
		if err != nil {
			slog.Error("Failed to marshal health response", "error", err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		if _, err := w.Write(jsonData); err != nil {
			slog.Error("Failed to write health response", "error", err)
		}
	})

	// Web UI
	mux.HandleFunc("/", s.handleWebUI)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// handleWebUI serves the web UI interface
func (s *Server) handleWebUI(w http.ResponseWriter, r *http.Request) {
	versionInfo := version.Get()
	metricsInfo := s.metrics.GetMetricsInfo()

	// Convert metrics to template data
	metrics := make([]MetricData, 0, len(metricsInfo))
	for _, metric := range metricsInfo {
		metrics = append(metrics, MetricData{
			Name:         metric.Name,
			Help:         metric.Help,
			Labels:       metric.Labels,
			ExampleValue: metric.ExampleValue,
		})
	}

	data := TemplateData{
		Version:   versionInfo.Version,
		Commit:    versionInfo.Commit,
		BuildDate: versionInfo.BuildDate,
		Status:    "Connected",
		Metrics:   metrics,
		Config: ConfigData{
			WebSocketURL: s.cfg.WebSocket.URL,
			DeviceCount:  0, // Hardcoded for now - would need Zigbee2MQTT client reference to get actual count
		},
	}

	w.Header().Set("Content-Type", "text/html")

	if err := mainTemplate.Execute(w, data); err != nil {
		http.Error(w, fmt.Sprintf("Error rendering template: %v", err), http.StatusInternalServerError)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	slog.Info("Starting server", "address", addr)

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}
