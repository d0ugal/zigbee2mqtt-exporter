package collectors

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// A trimmed bridge/health payload as published by Zigbee2MQTT 2.x.
const healthPayload = `{
	"topic": "bridge/health",
	"payload": {
		"response_time": 1794226450644,
		"os": {"load_average": [4.06, 4.32, 4.29], "memory_used_mb": 14763.93, "memory_percent": 46.9},
		"process": {"uptime_sec": 70805, "memory_used_mb": 200.5, "memory_percent": 0.63},
		"mqtt": {"connected": true, "queued": 2, "published": 154752, "received": 4857},
		"devices": {
			"0x9035eafffeb5fb49": {"messages": 122246, "messages_per_sec": 1.7265, "leave_count": 1, "network_address_changes": 3},
			"0x00000000deadbeef": {"messages": 5, "messages_per_sec": 0.1, "leave_count": 0, "network_address_changes": 0}
		}
	}
}`

const healthDevicesPayload = `{
	"topic": "bridge/devices",
	"payload": [
		{"ieee_address": "0x9035eafffeb5fb49", "friendly_name": "Hall Radiator", "type": "EndDevice"}
	]
}`

func TestProcessBridgeHealthMessage(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	collector.processMessage(ctx, []byte(healthDevicesPayload))
	collector.processMessage(ctx, []byte(healthPayload))

	m := registry.Health
	device := prometheus.Labels{"device": "Hall Radiator"}
	none := prometheus.Labels{}

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"device messages", testutil.ToFloat64(m.DeviceMessages.With(device)), 122246},
		{"device messages per second", testutil.ToFloat64(m.DeviceMessagesPerSecond.With(device)), 1.7265},
		{"device leave count", testutil.ToFloat64(m.DeviceLeaveCount.With(device)), 1},
		{"device network address changes", testutil.ToFloat64(m.DeviceNetworkAddressChanges.With(device)), 3},
		{"process uptime", testutil.ToFloat64(m.ProcessUptime.With(none)), 70805},
		{"process memory bytes", testutil.ToFloat64(m.ProcessMemoryUsed.With(none)), 200.5 * 1024 * 1024},
		{"mqtt connected", testutil.ToFloat64(m.MQTTConnected.With(none)), 1},
		{"mqtt queued", testutil.ToFloat64(m.MQTTQueued.With(none)), 2},
		{"mqtt published", testutil.ToFloat64(m.MQTTPublished.With(none)), 154752},
		{"mqtt received", testutil.ToFloat64(m.MQTTReceived.With(none)), 4857},
		{"last report", testutil.ToFloat64(m.LastReport.With(none)), 1794226450.644},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}

	// The device missing from bridge/devices must not produce a series.
	if n := testutil.CollectAndCount(m.DeviceMessages); n != 1 {
		t.Errorf("zigbee2mqtt_device_health_messages has %d series, want 1", n)
	}
}

func TestProcessBridgeHealthMessage_BeforeBridgeDevices(t *testing.T) {
	collector, registry := newTestCollector(t)

	collector.processMessage(context.Background(), []byte(healthPayload))

	if n := testutil.CollectAndCount(registry.Health.DeviceMessages); n != 0 {
		t.Errorf("expected no device series before bridge/devices, got %d", n)
	}

	// Bridge-wide stats do not depend on device names.
	if got := testutil.ToFloat64(registry.Health.MQTTConnected.With(prometheus.Labels{})); got != 1 {
		t.Errorf("mqtt connected = %v, want 1", got)
	}
}
