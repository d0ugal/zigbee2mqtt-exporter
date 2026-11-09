package collectors

import (
	"context"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// A trimmed bridge/info payload as published by Zigbee2MQTT 2.x with a Z-Stack coordinator.
const infoPayload = `{
	"topic": "bridge/info",
	"payload": {
		"version": "2.14.3",
		"commit": "3c841fdfd66c92239921ef84fde8f039c64e1662\n",
		"coordinator": {
			"ieee_address": "0x00124b002578e9e2",
			"type": "ZStack3x0",
			"meta": {"majorrel": 2, "minorrel": 7, "maintrel": 1, "product": 1, "revision": 20221226, "transportrev": 2}
		},
		"network": {"channel": 12, "pan_id": 6754, "extended_pan_id": "0x00124b002578e9e2"},
		"permit_join": true,
		"restart_required": false,
		"zigbee_herdsman": {"version": "11.2.0"},
		"zigbee_herdsman_converters": {"version": "26.125.0"}
	}
}`

func TestProcessBridgeInfoMessage(t *testing.T) {
	collector, registry := newTestCollector(t)
	collector.processMessage(context.Background(), []byte(infoPayload))

	m := registry.BridgeInfo
	none := prometheus.Labels{}

	expectedInfo := `
# HELP zigbee2mqtt_bridge_info Zigbee2MQTT version information (always 1, from bridge/info)
# TYPE zigbee2mqtt_bridge_info gauge
zigbee2mqtt_bridge_info{commit="3c841fdfd66c92239921ef84fde8f039c64e1662",version="2.14.3",zigbee_herdsman_converters_version="26.125.0",zigbee_herdsman_version="11.2.0"} 1
`
	if err := testutil.CollectAndCompare(m.Info, strings.NewReader(expectedInfo)); err != nil {
		t.Error(err)
	}

	expectedCoordinator := `
# HELP zigbee2mqtt_coordinator_info Zigbee coordinator adapter type and firmware revision (always 1, from bridge/info)
# TYPE zigbee2mqtt_coordinator_info gauge
zigbee2mqtt_coordinator_info{ieee_address="0x00124b002578e9e2",revision="20221226",type="ZStack3x0"} 1
`
	if err := testutil.CollectAndCompare(m.CoordinatorInfo, strings.NewReader(expectedCoordinator)); err != nil {
		t.Error(err)
	}

	checks := []struct {
		name string
		got  float64
		want float64
	}{
		{"channel", testutil.ToFloat64(m.NetworkChannel.With(none)), 12},
		{"pan id", testutil.ToFloat64(m.NetworkPanID.With(none)), 6754},
		{"permit join", testutil.ToFloat64(m.PermitJoin.With(none)), 1},
		{"restart required", testutil.ToFloat64(m.RestartRequired.With(none)), 0},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestProcessBridgeInfoMessage_UpgradeReplacesSeries(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	collector.processMessage(ctx, []byte(infoPayload))
	collector.processMessage(ctx, []byte(strings.Replace(infoPayload, `"version": "2.14.3"`, `"version": "2.15.0"`, 1)))

	if n := testutil.CollectAndCount(registry.BridgeInfo.Info); n != 1 {
		t.Fatalf("zigbee2mqtt_bridge_info has %d series after upgrade, want 1", n)
	}

	if got := testutil.ToFloat64(registry.BridgeInfo.Info.With(prometheus.Labels{
		"version":                            "2.15.0",
		"commit":                             "3c841fdfd66c92239921ef84fde8f039c64e1662",
		"zigbee_herdsman_version":            "11.2.0",
		"zigbee_herdsman_converters_version": "26.125.0",
	})); got != 1 {
		t.Errorf("new version series = %v, want 1", got)
	}
}

func TestFormatRevision(t *testing.T) {
	tests := []struct {
		in   interface{}
		want string
	}{
		{float64(20221226), "20221226"},
		{"7.4.4 [GA]", "7.4.4 [GA]"},
		{"", "unknown"},
		{nil, "unknown"},
	}

	for _, tt := range tests {
		if got := formatRevision(tt.in); got != tt.want {
			t.Errorf("formatRevision(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
