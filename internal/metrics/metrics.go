package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	return &Registry{
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
			[]string{"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state"},
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
	}
}
