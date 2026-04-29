package collectors

import (
	"context"
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
