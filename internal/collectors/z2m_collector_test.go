package collectors

import (
	"testing"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
)

func TestNewZ2MCollector(t *testing.T) {
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://localhost:8081/api",
		},
	}

	registry := metrics.NewZ2MRegistry(nil)

	// Test that NewZ2MCollector doesn't panic
	collector := NewZ2MCollector(cfg, registry)

	if collector == nil {
		t.Error("Expected collector to not be nil")
	}
}
