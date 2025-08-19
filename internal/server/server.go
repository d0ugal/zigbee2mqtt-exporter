package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

type MetricInfo struct {
	Name         string
	Help         string
	ExampleValue string
	Labels       map[string]string
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

	// Metrics info endpoint
	mux.HandleFunc("/metrics-info", s.handleMetricsInfo)

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

func (s *Server) getMetricsInfo() []MetricInfo {
	var metricsInfo []MetricInfo

	// Define all metrics manually since reflection approach is complex with Prometheus metrics
	metrics := []struct {
		name  string
		field string
	}{
		{"zigbee2mqtt_exporter_info", "VersionInfo"},
		{"zigbee2mqtt_device_last_seen_timestamp", "DeviceLastSeen"},
		{"zigbee2mqtt_device_seen_total", "DeviceSeenCount"},
		{"zigbee2mqtt_device_link_quality", "DeviceLinkQuality"},
		{"zigbee2mqtt_device_power_state", "DeviceState"},
		{"zigbee2mqtt_device_battery_level", "DeviceBattery"},
		{"zigbee2mqtt_device_info", "DeviceInfo"},
		{"zigbee2mqtt_device_up", "DeviceAvailability"},
		{"zigbee2mqtt_bridge_state", "BridgeState"},
		{"zigbee2mqtt_bridge_events_total", "BridgeEventsTotal"},
		{"zigbee2mqtt_websocket_connection_status", "WebSocketConnectionStatus"},
		{"zigbee2mqtt_websocket_messages_total", "WebSocketMessagesTotal"},
		{"zigbee2mqtt_websocket_reconnects_total", "WebSocketReconnectsTotal"},
		{"zigbee2mqtt_device_ota_update_available", "DeviceOTAUpdateAvailable"},
		{"zigbee2mqtt_device_current_firmware_version", "DeviceCurrentFirmware"},
		{"zigbee2mqtt_device_available_firmware_version", "DeviceAvailableFirmware"},
	}

	for _, metric := range metrics {
		metricsInfo = append(metricsInfo, MetricInfo{
			Name:         metric.name,
			Help:         s.getMetricHelp(metric.field),
			ExampleValue: s.getExampleValue(metric.field),
			Labels:       s.getExampleLabels(metric.field),
		})
	}

	return metricsInfo
}

func (s *Server) getExampleLabels(metricName string) map[string]string {
	switch metricName {
	case "VersionInfo":
		return map[string]string{"version": "v2.5.1", "commit": "abc123", "build_date": "2024-01-01"}
	case "DeviceLastSeen", "DeviceSeenCount", "DeviceLinkQuality", "DeviceState", "DeviceBattery", "DeviceAvailability", "DeviceOTAUpdateAvailable":
		return map[string]string{"device": "0x00158d0009b8b123"}
	case "DeviceInfo":
		return map[string]string{
			"device":            "0x00158d0009b8b123",
			"type":              "EndDevice",
			"power_source":      "Battery",
			"manufacturer":      "Xiaomi",
			"model_id":          "SNZB-02",
			"supported":         "true",
			"disabled":          "false",
			"interview_state":   "completed",
			"software_build_id": "3000-0001",
			"date_code":         "20201201",
		}
	case "BridgeState", "WebSocketConnectionStatus", "WebSocketReconnectsTotal":
		return map[string]string{}
	case "BridgeEventsTotal":
		return map[string]string{"event_type": "device_joined"}
	case "WebSocketMessagesTotal":
		return map[string]string{"topic": "zigbee2mqtt/bridge/devices"}
	case "DeviceCurrentFirmware", "DeviceAvailableFirmware":
		return map[string]string{"device": "0x00158d0009b8b123", "firmware_version": "1.0.0"}
	default:
		return map[string]string{}
	}
}

func (s *Server) getExampleValue(metricName string) string {
	switch metricName {
	case "VersionInfo":
		return "1"
	case "DeviceLastSeen":
		return "1704067200"
	case "DeviceSeenCount":
		return "42"
	case "DeviceLinkQuality":
		return "85"
	case "DeviceState":
		return "1"
	case "DeviceBattery":
		return "75"
	case "DeviceInfo":
		return "1"
	case "DeviceAvailability":
		return "1"
	case "BridgeState":
		return "1"
	case "BridgeEventsTotal":
		return "15"
	case "WebSocketConnectionStatus":
		return "1"
	case "WebSocketMessagesTotal":
		return "1024"
	case "WebSocketReconnectsTotal":
		return "3"
	case "DeviceOTAUpdateAvailable":
		return "0"
	case "DeviceCurrentFirmware":
		return "1"
	case "DeviceAvailableFirmware":
		return "1"
	default:
		return "0"
	}
}

func (s *Server) getMetricHelp(metricName string) string {
	switch metricName {
	case "VersionInfo":
		return "Information about the Zigbee2MQTT exporter"
	case "DeviceLastSeen":
		return "Timestamp when device was last seen"
	case "DeviceSeenCount":
		return "Number of times device has been seen"
	case "DeviceLinkQuality":
		return "Device link quality (0-255)"
	case "DeviceState":
		return "Device power state (1=ON, 0=OFF)"
	case "DeviceBattery":
		return "Device battery level (0-100)"
	case "DeviceInfo":
		return "Device information (always 1, used for joining with other metrics)"
	case "DeviceAvailability":
		return "Device availability status (1=online, 0=offline)"
	case "BridgeState":
		return "Bridge state (1=online, 0=offline)"
	case "BridgeEventsTotal":
		return "Total number of bridge events"
	case "WebSocketConnectionStatus":
		return "WebSocket connection status (1=connected, 0=disconnected)"
	case "WebSocketMessagesTotal":
		return "Total number of WebSocket messages received"
	case "WebSocketReconnectsTotal":
		return "Total number of WebSocket reconnections"
	case "DeviceOTAUpdateAvailable":
		return "Device OTA update availability (1=available, 0=not_available)"
	case "DeviceCurrentFirmware":
		return "Device current firmware version (always 1, used for joining with other metrics)"
	case "DeviceAvailableFirmware":
		return "Device available firmware version (always 1, used for joining with other metrics)"
	default:
		return "Zigbee2MQTT exporter metric"
	}
}

func (s *Server) handleMetricsInfo(w http.ResponseWriter, r *http.Request) {
	metricsInfo := s.getMetricsInfo()

	response := map[string]interface{}{
		"metrics":      metricsInfo,
		"total_count":  len(metricsInfo),
		"generated_at": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal metrics info response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	if _, err := w.Write(jsonData); err != nil {
		slog.Error("Failed to write metrics info response", "error", err)
	}
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
	metricsInfo := s.getMetricsInfo()

	// Generate metrics HTML dynamically
	metricsHTML := ""

	for i, metric := range metricsInfo {
		labelsStr := ""

		if len(metric.Labels) > 0 {
			var labelPairs []string
			for k, v := range metric.Labels {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, k, v))
			}

			labelsStr = "{" + strings.Join(labelPairs, ", ") + "}"
		}

		// Create clickable metric with hidden details
		metricsHTML += fmt.Sprintf(`
            <div class="metric-item" onclick="toggleMetricDetails(%d)">
                <div class="metric-header">
                    <span class="metric-name">%s</span>
                    <span class="metric-toggle">▼</span>
                </div>
                <div class="metric-details" id="metric-%d">
                    <div class="metric-help"><strong>Description:</strong> %s</div>
                    <div class="metric-example"><strong>Example:</strong> %s = %s</div>
                    <div class="metric-labels"><strong>Labels:</strong> %s</div>
                </div>
            </div>`,
			i,
			metric.Name,
			i,
			metric.Help,
			metric.Name,
			metric.ExampleValue,
			labelsStr)
	}

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
        .metrics-list {
            margin: 0.5rem 0;
        }
        .metric-item {
            border: 1px solid #dee2e6;
            border-radius: 6px;
            margin: 0.5rem 0;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        .metric-item:hover {
            border-color: #007bff;
            background-color: #f8f9fa;
        }
        .metric-header {
            padding: 0.75rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-weight: 500;
            color: #495057;
        }
        .metric-name {
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
        }
        .metric-toggle {
            font-size: 0.8rem;
            color: #6c757d;
            transition: transform 0.2s ease;
        }
        .metric-details {
            display: none;
            padding: 0.75rem;
            border-top: 1px solid #dee2e6;
            background-color: #f8f9fa;
            font-size: 0.85rem;
            line-height: 1.4;
        }
        .metric-details.show {
            display: block;
        }
        .metric-help, .metric-example, .metric-labels {
            margin: 0.5rem 0;
        }
        .metric-example {
            font-family: 'Courier New', monospace;
            background-color: #e9ecef;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
        }
        .metric-labels {
            color: #6c757d;
        }
    </style>
    <script>
        function toggleMetricDetails(id) {
            const details = document.getElementById('metric-' + id);
            const toggle = details.previousElementSibling.querySelector('.metric-toggle');
            
            if (details.classList.contains('show')) {
                details.classList.remove('show');
                toggle.textContent = '▼';
            } else {
                details.classList.add('show');
                toggle.textContent = '▲';
            }
        }
    </script>
</head>
<body>
    <h1>Zigbee2MQTT Exporter<span class="version">` + versionInfo.Version + `</span></h1>
    
    <div class="endpoint">
        <h3><a href="/metrics">📊 Metrics</a></h3>
        <p class="description">Prometheus metrics endpoint</p>
        <span class="status metrics">Available</span>
    </div>

    <div class="endpoint">
        <h3><a href="/metrics-info">📋 Metrics Info</a></h3>
        <p class="description">Detailed metrics information with examples</p>
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
        <div class="metrics-list">` + metricsHTML + `
        </div>
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
