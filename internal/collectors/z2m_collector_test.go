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
		name          string
		otaState      string
		expectedGauge float64
	}{
		{"idle means no update", "idle", 0.0},
		{"available means update pending", "available", 1.0},
		{"scheduled means update pending", "scheduled", 1.0},
		{"updating means update in progress", "updating", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector, registry := newTestCollector(t)

			payload := map[string]interface{}{
				"update": map[string]interface{}{
					"state":             tt.otaState,
					"installed_version": float64(620834817),
					"latest_version":    float64(620834817),
				},
			}

			collector.updateDeviceMetrics(context.Background(), "test_device", payload)

			got := testutil.ToFloat64(registry.DeviceOTAUpdateAvailable.With(prometheus.Labels{
				"device": "test_device",
			}))

			if got != tt.expectedGauge {
				t.Errorf("OTA state %q: got %v, want %v", tt.otaState, got, tt.expectedGauge)
			}
		})
	}
}
