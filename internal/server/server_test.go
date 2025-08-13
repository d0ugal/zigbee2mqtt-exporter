package server

import (
	"testing"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			Host: "127.0.0.1",
			Port: 8087,
		},
	}

	registry := metrics.NewRegistry()

	// Test that New doesn't panic
	server := New(cfg, registry)

	if server == nil {
		t.Error("Expected server to not be nil")
	}
}
