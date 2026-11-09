package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HealthMetrics holds the metrics populated from Zigbee2MQTT's bridge/health
// topic. Zigbee2MQTT owns the underlying counters (they reset when it
// restarts), so they are mirrored as gauges; rate() and increase() still work.
type HealthMetrics struct {
	DeviceMessages              *prometheus.GaugeVec
	DeviceMessagesPerSecond     *prometheus.GaugeVec
	DeviceLeaveCount            *prometheus.GaugeVec
	DeviceNetworkAddressChanges *prometheus.GaugeVec

	ProcessUptime     *prometheus.GaugeVec
	ProcessMemoryUsed *prometheus.GaugeVec

	MQTTConnected *prometheus.GaugeVec
	MQTTQueued    *prometheus.GaugeVec
	MQTTPublished *prometheus.GaugeVec
	MQTTReceived  *prometheus.GaugeVec

	LastReport *prometheus.GaugeVec
}

func newHealthMetrics(baseRegistry *promexporter_metrics.Registry, factory promauto.Factory) *HealthMetrics {
	gauge := func(name, help string, labels []string) *prometheus.GaugeVec {
		baseRegistry.AddMetricInfo(name, help, labels)

		return factory.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labels)
	}

	device := []string{"device"}
	none := []string{}

	return &HealthMetrics{
		DeviceMessages: gauge("zigbee2mqtt_device_health_messages",
			"Messages received from the device since Zigbee2MQTT started (from bridge/health)", device),
		DeviceMessagesPerSecond: gauge("zigbee2mqtt_device_health_messages_per_second",
			"Average messages per second received from the device since Zigbee2MQTT started (from bridge/health)", device),
		DeviceLeaveCount: gauge("zigbee2mqtt_device_health_leave_count",
			"Times the device has left the network since Zigbee2MQTT started (from bridge/health)", device),
		DeviceNetworkAddressChanges: gauge("zigbee2mqtt_device_health_network_address_changes",
			"Times the device's network address has changed since Zigbee2MQTT started (from bridge/health)", device),

		ProcessUptime: gauge("zigbee2mqtt_process_uptime_seconds",
			"Zigbee2MQTT process uptime in seconds (from bridge/health)", none),
		ProcessMemoryUsed: gauge("zigbee2mqtt_process_memory_used_bytes",
			"Zigbee2MQTT process memory usage in bytes (from bridge/health)", none),

		MQTTConnected: gauge("zigbee2mqtt_mqtt_connected",
			"Whether Zigbee2MQTT is connected to the MQTT broker (1=connected, 0=disconnected, from bridge/health)", none),
		MQTTQueued: gauge("zigbee2mqtt_mqtt_queued_messages",
			"MQTT messages queued by Zigbee2MQTT waiting to be published (from bridge/health)", none),
		MQTTPublished: gauge("zigbee2mqtt_mqtt_published_messages",
			"MQTT messages published by Zigbee2MQTT since it started (from bridge/health)", none),
		MQTTReceived: gauge("zigbee2mqtt_mqtt_received_messages",
			"MQTT messages received by Zigbee2MQTT since it started (from bridge/health)", none),

		LastReport: gauge("zigbee2mqtt_health_last_report_timestamp_seconds",
			"Unix timestamp of the most recent bridge/health report", none),
	}
}
