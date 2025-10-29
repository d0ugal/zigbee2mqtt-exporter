package collectors

import (
	"testing"

	"github.com/d0ugal/promexporter/app"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
)

func TestNewZ2MCollector(t *testing.T) {
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://localhost:8081/api",
		},
	}

	// Create a mock base registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)

	// Create a minimal app instance for testing
	testApp := app.New("Test Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(baseRegistry).
		Build()

	// Test that NewZ2MCollector doesn't panic
	collector := NewZ2MCollector(cfg, registry, testApp)

	if collector == nil {
		t.Error("Expected collector to not be nil")
	}
}
