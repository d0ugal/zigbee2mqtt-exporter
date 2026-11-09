package collectors

import (
	"context"
	"strings"
	"testing"

	"github.com/d0ugal/promexporter/app"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestCollector(t *testing.T) (*Z2MCollector, *metrics.Z2MRegistry) {
	t.Helper()

	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://localhost:8081/api",
		},
	}

	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)
	testApp := app.New("Test Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(baseRegistry).
		Build()

	return NewZ2MCollector(cfg, registry, testApp), registry
}

func TestNewZ2MCollector(t *testing.T) {
	collector, _ := newTestCollector(t)

	if collector == nil {
		t.Error("Expected collector to not be nil")
	}
}

// TestDeviceFirmware_LabelsMatchRegistry guards against the label-name
// mismatch between the registry definitions (device, firmware_version) and
// the call sites in processBridgeDevicesMessage. If the labels drift again,
// prometheus.Labels.With() panics here.
func TestDeviceFirmware_LabelsMatchRegistry(t *testing.T) {
	_, registry := newTestCollector(t)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	registry.DeviceCurrentFirmware.With(prometheus.Labels{
		"device":           "test-device",
		"firmware_version": "1.2.3",
	}).Set(1)
	registry.DeviceAvailableFirmware.With(prometheus.Labels{
		"device":           "test-device",
		"firmware_version": "1.2.4",
	}).Set(1)
}

func TestUpdateDeviceMetrics_OTAState(t *testing.T) {
	tests := []struct {
		name             string
		otaState         string
		installedVersion float64
		latestVersion    float64
		wantIdle         float64
		wantAvail        float64
		wantSched        float64
		wantUpd          float64
	}{
		{"idle same versions", "idle", 620834817, 620834817, 1, 0, 0, 0},
		{"idle newer version available", "idle", 620834817, 620834818, 0, 1, 0, 0},
		{"idle installed newer than latest", "idle", 620834818, 620834817, 1, 0, 0, 0},
		{"available", "available", 620834817, 620834817, 0, 1, 0, 0},
		{"scheduled", "scheduled", 620834817, 620834817, 0, 0, 1, 0},
		{"updating", "updating", 620834817, 620834817, 0, 0, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, registry := newTestCollector(t)

			payload := map[string]interface{}{
				"update": map[string]interface{}{
					"state":             tt.otaState,
					"installed_version": tt.installedVersion,
					"latest_version":    tt.latestVersion,
				},
			}

			collector.updateDeviceMetrics(context.Background(), "test_device", payload)

			check := func(state string, want float64) {
				t.Helper()

				labels := prometheus.Labels{"device": "test_device", "state": state}
				if got := testutil.ToFloat64(registry.DeviceOTAState.With(labels)); got != want {
					t.Errorf("OTA state %q: zigbee2mqtt_device_ota{state=%q} = %v, want %v", tt.otaState, state, got, want)
				}
			}

			check("idle", tt.wantIdle)
			check("available", tt.wantAvail)
			check("scheduled", tt.wantSched)
			check("updating", tt.wantUpd)
		})
	}
}

func TestUpdateDeviceMetrics_FirmwareVersions(t *testing.T) {
	collector, registry := newTestCollector(t)
	ctx := context.Background()

	update := func(installed, latest float64) {
		collector.updateDeviceMetrics(ctx, "test_device", map[string]interface{}{
			"update": map[string]interface{}{
				"state":             "idle",
				"installed_version": installed,
				"latest_version":    latest,
			},
		})
	}

	// Real OTA file versions are large integers; they must not be rendered
	// in exponent form.
	update(620834817, 620834818)

	expected := `
# HELP zigbee2mqtt_device_current_firmware_version Device current firmware version (always 1, used for joining with other metrics)
# TYPE zigbee2mqtt_device_current_firmware_version gauge
zigbee2mqtt_device_current_firmware_version{device="test_device",firmware_version="620834817"} 1
`
	if err := testutil.CollectAndCompare(registry.DeviceCurrentFirmware, strings.NewReader(expected)); err != nil {
		t.Error(err)
	}

	if got := testutil.ToFloat64(registry.DeviceAvailableFirmware.With(prometheus.Labels{
		"device": "test_device", "firmware_version": "620834818",
	})); got != 1 {
		t.Errorf("available firmware = %v, want 1", got)
	}

	// After the upgrade the old version series must be gone.
	update(620834818, 620834818)

	expected = `
# HELP zigbee2mqtt_device_current_firmware_version Device current firmware version (always 1, used for joining with other metrics)
# TYPE zigbee2mqtt_device_current_firmware_version gauge
zigbee2mqtt_device_current_firmware_version{device="test_device",firmware_version="620834818"} 1
`
	if err := testutil.CollectAndCompare(registry.DeviceCurrentFirmware, strings.NewReader(expected)); err != nil {
		t.Error(err)
	}

	if n := testutil.CollectAndCount(registry.DeviceAvailableFirmware); n != 1 {
		t.Errorf("available firmware has %d series, want 1", n)
	}
}

func TestUpdateDeviceMetrics_FirmwareVersionsMissing(t *testing.T) {
	collector, registry := newTestCollector(t)

	// A device that has never been checked for updates has no versions.
	collector.updateDeviceMetrics(context.Background(), "test_device", map[string]interface{}{
		"update": map[string]interface{}{"state": "idle", "installed_version": nil, "latest_version": nil},
	})

	if n := testutil.CollectAndCount(registry.DeviceCurrentFirmware); n != 0 {
		t.Errorf("current firmware has %d series, want 0", n)
	}
}
