package logging

import (
	"testing"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
)

func TestConfigure(t *testing.T) {
	// Test that Configure doesn't panic
	loggingConfig := &config.LoggingConfig{
		Level:  "info",
		Format: "json",
	}

	// This should not panic
	Configure(loggingConfig)
}
