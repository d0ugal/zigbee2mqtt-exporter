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

// Start starts the HTTP server
func (s *Server) Start() error {
	slog.Info("Starting HTTP server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	slog.Info("Shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// handleWebUI serves the web UI interface
func (s *Server) handleWebUI(w http.ResponseWriter, r *http.Request) {
	versionInfo := version.Get()
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zigbee2MQTT Exporter ` + versionInfo.Version + `</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            line-height: 1.6;
            color: #333;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 0.5rem;
        }
        h1 .version {
            font-size: 0.6em;
            color: #6c757d;
            font-weight: normal;
            margin-left: 0.5rem;
        }
        .endpoint {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .endpoint h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .endpoint a {
            color: #007bff;
            text-decoration: none;
            font-weight: 500;
        }
        .endpoint a:hover {
            text-decoration: underline;
        }
        .description {
            color: #6c757d;
            font-size: 0.9rem;
        }
        .status {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8rem;
            font-weight: 500;
        }
        .status.healthy {
            background: #d4edda;
            color: #155724;
        }
        .status.metrics {
            background: #d1ecf1;
            color: #0c5460;
        }
        .status.ready {
            background: #d4edda;
            color: #155724;
        }
        .status.connected {
            background: #d4edda;
            color: #155724;
        }
        .status.disconnected {
            background: #f8d7da;
            color: #721c24;
        }
        .service-status {
            background: #e9ecef;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .service-status h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .service-status p {
            margin: 0.25rem 0;
            color: #6c757d;
        }
        .metrics-info {
            background: #e9ecef;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .metrics-info h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .metrics-info ul {
            margin: 0.5rem 0;
            padding-left: 1.5rem;
        }
        .metrics-info li {
            margin: 0.25rem 0;
            color: #6c757d;
        }
        .footer {
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid #dee2e6;
            text-align: center;
            color: #6c757d;
            font-size: 0.9rem;
        }
        .footer a {
            color: #007bff;
            text-decoration: none;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>Zigbee2MQTT Exporter<span class="version">` + versionInfo.Version + `</span></h1>
    
    <div class="endpoint">
        <h3><a href="/metrics">📊 Metrics</a></h3>
        <p class="description">Prometheus metrics endpoint</p>
        <span class="status metrics">Available</span>
    </div>

    <div class="endpoint">
        <h3><a href="/health">❤️ Health Check</a></h3>
        <p class="description">Service health status</p>
        <span class="status healthy">Healthy</span>
    </div>

    <div class="service-status">
        <h3>Service Status</h3>
        <p><strong>Status:</strong> <span class="status ready">Ready</span></p>
        <p><strong>WebSocket Connection:</strong> <span class="status connected">Connected</span></p>
        <p><strong>Device Monitoring:</strong> <span class="status ready">Active</span></p>
    </div>

    <div class="metrics-info">
        <h3>Version Information</h3>
        <ul>
            <li><strong>Version:</strong> ` + versionInfo.Version + `</li>
            <li><strong>Commit:</strong> ` + versionInfo.Commit + `</li>
            <li><strong>Build Date:</strong> ` + versionInfo.BuildDate + `</li>
        </ul>
    </div>

    <div class="metrics-info">
        <h3>Configuration</h3>
        <ul>
            <li><strong>WebSocket URL:</strong> ws://localhost:8081/api</li>
            <li><strong>Server Port:</strong> 8087</li>
            <li><strong>Log Level:</strong> info</li>
        </ul>
    </div>

    <div class="metrics-info">
        <h3>Available Metrics</h3>
        <ul>
            <li><strong>zigbee2mqtt_device_last_seen_timestamp:</strong> Timestamp when device was last seen</li>
            <li><strong>zigbee2mqtt_device_seen_count:</strong> Number of times device has been seen</li>
            <li><strong>zigbee2mqtt_device_link_quality:</strong> Device link quality (0-255)</li>
            <li><strong>zigbee2mqtt_device_power_state:</strong> Device power state (1=ON, 0=OFF)</li>
            <li><strong>zigbee2mqtt_device_battery_level:</strong> Device battery level (0-100)</li>
            <li><strong>zigbee2mqtt_device_info:</strong> Device information and metadata</li>
            <li><strong>zigbee2mqtt_device_up:</strong> Device availability status</li>
            <li><strong>zigbee2mqtt_websocket_connection_status:</strong> WebSocket connection status</li>
            <li><strong>zigbee2mqtt_websocket_messages_total:</strong> Total messages received per topic</li>
            <li><strong>zigbee2mqtt_websocket_reconnects_total:</strong> Total reconnection attempts</li>
        </ul>
    </div>

    <div class="footer">
        <p>Copyright © 2024 zigbee2mqtt-exporter contributors. Licensed under <a href="https://opensource.org/licenses/MIT" target="_blank">MIT License</a>.</p>
        <p><a href="https://github.com/d0ugal/zigbee2mqtt-exporter" target="_blank">GitHub Repository</a> | <a href="https://github.com/d0ugal/zigbee2mqtt-exporter/issues" target="_blank">Report Issues</a></p>
    </div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if _, err := w.Write([]byte(html)); err != nil {
		slog.Error("Failed to write HTML response", "error", err)
	}
}
