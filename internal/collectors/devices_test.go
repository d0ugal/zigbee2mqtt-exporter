package collectors

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func devicesMessage(devices string) []byte {
	return []byte(`{"topic": "bridge/devices", "payload": [` + devices + `]}`)
}

const (
	kitchenV1 = `{"ieee_address": "0x01", "friendly_name": "Kitchen", "type": "Router", "software_build_id": "1.0"}`
	kitchenV2 = `{"ieee_address": "0x01", "friendly_name": "Kitchen", "type": "Router", "software_build_id": "2.0"}`
	hall      = `{"ieee_address": "0x02", "friendly_name": "Hall", "type": "EndDevice"}`
	hallNew   = `{"ieee_address": "0x02", "friendly_name": "Landing", "type": "EndDevice"}`
)

func TestProcessBridgeDevices_LabelChangeReplacesInfoSeries(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	collector.processMessage(ctx, devicesMessage(kitchenV1))
	collector.processMessage(ctx, devicesMessage(kitchenV2))

	if n := testutil.CollectAndCount(registry.DeviceInfo); n != 1 {
		t.Fatalf("zigbee2mqtt_device_info has %d series, want 1", n)
	}

	if got := testutil.ToFloat64(registry.DeviceInfo.With(prometheus.Labels{
		"device": "Kitchen", "type": "Router", "power_source": "unknown", "manufacturer": "",
		"model_id": "", "supported": "false", "disabled": "false", "interview_state": "not_started",
		"software_build_id": "2.0", "date_code": "",
	})); got != 1 {
		t.Errorf("new info series = %v, want 1", got)
	}
}

func TestProcessBridgeDevices_RenameRemovesOldSeries(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	collector.processMessage(ctx, devicesMessage(kitchenV1+","+hall))

	// Give both devices state, availability and health series.
	for _, name := range []string{"Kitchen", "Hall"} {
		collector.updateDeviceMetrics(ctx, name, map[string]interface{}{"linkquality": float64(100)})
		registry.DeviceAvailability.With(prometheus.Labels{"device": name}).Set(1)
	}

	collector.processMessage(ctx, []byte(`{"topic": "bridge/health", "payload": {"devices": {
		"0x01": {"messages": 1}, "0x02": {"messages": 2}}}}`))

	// A group publishes state too but never appears in bridge/devices.
	collector.updateDeviceMetrics(ctx, "Kitchen Group", map[string]interface{}{"state": "ON"})

	// Hall is renamed to Landing.
	collector.processMessage(ctx, devicesMessage(kitchenV1+","+hallNew))

	for _, tc := range []struct {
		name string
		vec  prometheus.Collector
		want int
	}{
		{"link_quality", registry.DeviceLinkQuality, 1},
		{"up", registry.DeviceAvailability, 1},
		{"health_messages", registry.Health.DeviceMessages, 1},
		// Kitchen plus the group
		{"seen_total", registry.DeviceSeenCount, 2},
	} {
		if n := testutil.CollectAndCount(tc.vec); n != tc.want {
			t.Errorf("%s has %d series after rename, want %d", tc.name, n, tc.want)
		}
	}

	if got := testutil.ToFloat64(registry.DeviceLinkQuality.With(prometheus.Labels{"device": "Kitchen"})); got != 100 {
		t.Errorf("Kitchen link quality = %v, want 100", got)
	}

	if n := testutil.CollectAndCount(registry.DeviceState); n != 1 {
		t.Errorf("group power state series removed; got %d series, want 1", n)
	}

	// Health for the renamed device now lands on the new name.
	collector.processMessage(ctx, []byte(`{"topic": "bridge/health", "payload": {"devices": {"0x02": {"messages": 3}}}}`))

	if got := testutil.ToFloat64(registry.Health.DeviceMessages.With(prometheus.Labels{"device": "Landing"})); got != 3 {
		t.Errorf("Landing health messages = %v, want 3", got)
	}
}

func TestProcessBridgeDevices_InvalidPayloadKeepsState(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	collector.processMessage(ctx, devicesMessage(kitchenV1))
	collector.processMessage(ctx, []byte(`{"topic": "bridge/devices", "payload": "not a list"}`))

	if n := testutil.CollectAndCount(registry.DeviceInfo); n != 1 {
		t.Errorf("zigbee2mqtt_device_info has %d series after bad payload, want 1", n)
	}
}
