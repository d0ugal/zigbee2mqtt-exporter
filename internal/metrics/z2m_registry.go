package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Z2MRegistry wraps the promexporter registry with Z2M-specific metrics
type Z2MRegistry struct {
	*promexporter_metrics.Registry

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
}

// NewZ2MRegistry creates a new Z2M metrics registry
func NewZ2MRegistry(baseRegistry *promexporter_metrics.Registry) *Z2MRegistry {
	// Get the underlying Prometheus registry
	promRegistry := baseRegistry.GetRegistry()
	factory := promauto.With(promRegistry)

	z2m := &Z2MRegistry{
		Registry: baseRegistry,
	}

	// Device metrics
	z2m.DeviceLastSeen = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_last_seen_timestamp",
			Help: "Timestamp when device was last seen",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_last_seen_timestamp", "Timestamp when device was last seen", []string{"device"})

	z2m.DeviceSeenCount = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zigbee2mqtt_device_seen_total",
			Help: "Number of times device has been seen",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_seen_total", "Number of times device has been seen", []string{"device"})

	z2m.DeviceLinkQuality = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_link_quality",
			Help: "Device link quality (0-255)",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_link_quality", "Device link quality (0-255)", []string{"device"})

	z2m.DeviceState = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_power_state",
			Help: "Device power state (1=ON, 0=OFF)",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_power_state", "Device power state (1=ON, 0=OFF)", []string{"device"})

	z2m.DeviceBattery = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_battery_level",
			Help: "Device battery level (0-100)",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_battery_level", "Device battery level (0-100)", []string{"device"})

	// Device info metric (always set to 1, used for joining)
	z2m.DeviceInfo = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_info",
			Help: "Device information (always 1, used for joining with other metrics)",
		},
		[]string{"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state", "software_build_id", "date_code"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_info", "Device information (always 1, used for joining with other metrics)", []string{"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state", "software_build_id", "date_code"})

	// Device availability metric (like Prometheus "up" metric)
	z2m.DeviceAvailability = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_up",
			Help: "Device availability status (1=online, 0=offline)",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_up", "Device availability status (1=online, 0=offline)", []string{"device"})

	// Bridge metrics
	z2m.BridgeState = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_bridge_state",
			Help: "Bridge state (1=online, 0=offline)",
		},
		[]string{},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_bridge_state", "Bridge state (1=online, 0=offline)", []string{})

	z2m.BridgeEventsTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zigbee2mqtt_bridge_events_total",
			Help: "Total number of bridge events",
		},
		[]string{"event_type"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_bridge_events_total", "Total number of bridge events", []string{"event_type"})

	// Connection metrics
	z2m.WebSocketConnectionStatus = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_websocket_connection_status",
			Help: "WebSocket connection status (1=connected, 0=disconnected)",
		},
		[]string{},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_websocket_connection_status", "WebSocket connection status (1=connected, 0=disconnected)", []string{})

	z2m.WebSocketMessagesTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zigbee2mqtt_websocket_messages_total",
			Help: "Total number of WebSocket messages received",
		},
		[]string{"topic"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_websocket_messages_total", "Total number of WebSocket messages received", []string{"topic"})

	z2m.WebSocketReconnectsTotal = factory.NewCounterVec(
		prometheus.CounterOpts{
			Name: "zigbee2mqtt_websocket_reconnects_total",
			Help: "Total number of WebSocket reconnections",
		},
		[]string{},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_websocket_reconnects_total", "Total number of WebSocket reconnections", []string{})

	// OTA Update metrics
	z2m.DeviceOTAUpdateAvailable = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_ota_update_available",
			Help: "Device OTA update availability (1=available, 0=not_available)",
		},
		[]string{"device"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_ota_update_available", "Device OTA update availability (1=available, 0=not_available)", []string{"device"})

	z2m.DeviceCurrentFirmware = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_current_firmware_version",
			Help: "Device current firmware version (always 1, used for joining with other metrics)",
		},
		[]string{"device", "firmware_version"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_current_firmware_version", "Device current firmware version (always 1, used for joining with other metrics)", []string{"device", "firmware_version"})

	z2m.DeviceAvailableFirmware = factory.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "zigbee2mqtt_device_available_firmware_version",
			Help: "Device available firmware version (always 1, used for joining with other metrics)",
		},
		[]string{"device", "firmware_version"},
	)
	baseRegistry.AddMetricInfo("zigbee2mqtt_device_available_firmware_version", "Device available firmware version (always 1, used for joining with other metrics)", []string{"device", "firmware_version"})

	return z2m
}
