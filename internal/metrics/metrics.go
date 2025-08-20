package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricInfo contains information about a metric for the UI
type MetricInfo struct {
	Name         string
	Help         string
	Labels       []string
	ExampleValue string
}

// Registry holds all the Prometheus metrics
type Registry struct {
	// Version info metric
	VersionInfo *prometheus.GaugeVec

	// Device metrics
	DeviceLastSeen    *prometheus.GaugeVec
	DeviceSeenCount   *prometheus.CounterVec
	DeviceLinkQuality *prometheus.GaugeVec
	DeviceState       *prometheus.GaugeVec
	DeviceBattery     *prometheus.GaugeVec

	// Device info metric (for joining with other metrics)
	DeviceInfo *prometheus.GaugeVec

	// Device availability metric (like Prometheus "up" metric)
	DeviceAvailability *prometheus.GaugeVec

	// Bridge metrics
	BridgeState       *prometheus.GaugeVec
	BridgeEventsTotal *prometheus.CounterVec

	// Connection metrics
	WebSocketConnectionStatus *prometheus.GaugeVec
	WebSocketMessagesTotal    *prometheus.CounterVec
	WebSocketReconnectsTotal  *prometheus.CounterVec

	// OTA Update metrics
	DeviceOTAUpdateAvailable *prometheus.GaugeVec
	DeviceCurrentFirmware    *prometheus.GaugeVec
	DeviceAvailableFirmware  *prometheus.GaugeVec

	// Metric information for UI
	metricInfo []MetricInfo
}

// addMetricInfo adds metric information to the registry
func (r *Registry) addMetricInfo(name, help string, labels []string) {
	r.metricInfo = append(r.metricInfo, MetricInfo{
		Name:         name,
		Help:         help,
		Labels:       labels,
		ExampleValue: "",
	})
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	r := &Registry{
		// Version info metric
		VersionInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_exporter_info",
				Help: "Information about the Zigbee2MQTT exporter",
			},
			[]string{"version", "commit", "build_date"},
		),
		// Device metrics
		DeviceLastSeen: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_last_seen_timestamp",
				Help: "Timestamp when device was last seen",
			},
			[]string{"device"},
		),
		DeviceSeenCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zigbee2mqtt_device_seen_total",
				Help: "Number of times device has been seen",
			},
			[]string{"device"},
		),
		DeviceLinkQuality: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_link_quality",
				Help: "Device link quality (0-255)",
			},
			[]string{"device"},
		),
		DeviceState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_power_state",
				Help: "Device power state (1=ON, 0=OFF)",
			},
			[]string{"device"},
		),
		DeviceBattery: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_battery_level",
				Help: "Device battery level (0-100)",
			},
			[]string{"device"},
		),

		// Device info metric (always set to 1, used for joining)
		DeviceInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_info",
				Help: "Device information (always 1, used for joining with other metrics)",
			},
			[]string{"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state", "software_build_id", "date_code"},
		),

		// Device availability metric (like Prometheus "up" metric)
		DeviceAvailability: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_up",
				Help: "Device availability status (1=online, 0=offline)",
			},
			[]string{"device"},
		),

		// Bridge metrics
		BridgeState: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_bridge_state",
				Help: "Bridge state (1=online, 0=offline)",
			},
			[]string{},
		),
		BridgeEventsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zigbee2mqtt_bridge_events_total",
				Help: "Total number of bridge events",
			},
			[]string{"event_type"},
		),

		// Connection metrics
		WebSocketConnectionStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_websocket_connection_status",
				Help: "WebSocket connection status (1=connected, 0=disconnected)",
			},
			[]string{},
		),
		WebSocketMessagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zigbee2mqtt_websocket_messages_total",
				Help: "Total number of WebSocket messages received",
			},
			[]string{"topic"},
		),
		WebSocketReconnectsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "zigbee2mqtt_websocket_reconnects_total",
				Help: "Total number of WebSocket reconnections",
			},
			[]string{},
		),

		// OTA Update metrics
		DeviceOTAUpdateAvailable: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_ota_update_available",
				Help: "Device OTA update availability (1=available, 0=not_available)",
			},
			[]string{"device"},
		),
		DeviceCurrentFirmware: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_current_firmware_version",
				Help: "Device current firmware version (always 1, used for joining with other metrics)",
			},
			[]string{"device", "firmware_version"},
		),
		DeviceAvailableFirmware: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "zigbee2mqtt_device_available_firmware_version",
				Help: "Device available firmware version (always 1, used for joining with other metrics)",
			},
			[]string{"device", "firmware_version"},
		),
	}

	// Add metric information for UI
	r.addMetricInfo("zigbee2mqtt_exporter_info", "Information about the Zigbee2MQTT exporter", []string{"version", "commit", "build_date"})
	r.addMetricInfo("zigbee2mqtt_device_last_seen_timestamp", "Timestamp when device was last seen", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_seen_total", "Number of times device has been seen", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_link_quality", "Device link quality (0-255)", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_power_state", "Device power state (1=ON, 0=OFF)", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_battery_level", "Device battery level (0-100)", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_info", "Device information (always 1, used for joining with other metrics)", []string{"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state", "software_build_id", "date_code"})
	r.addMetricInfo("zigbee2mqtt_device_up", "Device availability status (1=online, 0=offline)", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_bridge_state", "Bridge state (1=online, 0=offline)", []string{})
	r.addMetricInfo("zigbee2mqtt_bridge_events_total", "Total number of bridge events", []string{"event_type"})
	r.addMetricInfo("zigbee2mqtt_websocket_connection_status", "WebSocket connection status (1=connected, 0=disconnected)", []string{})
	r.addMetricInfo("zigbee2mqtt_websocket_messages_total", "Total number of WebSocket messages received", []string{"topic"})
	r.addMetricInfo("zigbee2mqtt_websocket_reconnects_total", "Total number of WebSocket reconnections", []string{})
	r.addMetricInfo("zigbee2mqtt_device_ota_update_available", "Device OTA update availability (1=available, 0=not_available)", []string{"device"})
	r.addMetricInfo("zigbee2mqtt_device_current_firmware_version", "Device current firmware version (always 1, used for joining with other metrics)", []string{"device", "firmware_version"})
	r.addMetricInfo("zigbee2mqtt_device_available_firmware_version", "Device available firmware version (always 1, used for joining with other metrics)", []string{"device", "firmware_version"})

	return r
}

// GetMetricsInfo returns information about all metrics for the UI
func (r *Registry) GetMetricsInfo() []MetricInfo {
	return r.metricInfo
}
