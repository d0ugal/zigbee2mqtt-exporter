package config

import (
	"os"
	"strconv"

	"github.com/d0ugal/promexporter/config"
)

// Config represents the application configuration
type Config struct {
	config.BaseConfig

	WebSocket WebSocketConfig `yaml:"websocket"`
}

// WebSocketConfig represents the WebSocket configuration
type WebSocketConfig struct {
	URL string `yaml:"url"`
}

// LoadFromEnvironment loads configuration from environment variables
func LoadFromEnvironment() *Config {
	cfg := &Config{}

	// Load base configuration from environment
	baseConfig := &config.BaseConfig{}

	// Server configuration
	if host := os.Getenv("Z2M_EXPORTER_SERVER_HOST"); host != "" {
		baseConfig.Server.Host = host
	} else {
		baseConfig.Server.Host = "0.0.0.0"
	}

	if portStr := os.Getenv("Z2M_EXPORTER_SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			baseConfig.Server.Port = port
		} else {
			baseConfig.Server.Port = 8087
		}
	} else {
		baseConfig.Server.Port = 8087
	}

	// Logging configuration
	if level := os.Getenv("Z2M_EXPORTER_LOG_LEVEL"); level != "" {
		baseConfig.Logging.Level = level
	} else {
		baseConfig.Logging.Level = "info"
	}

	if format := os.Getenv("Z2M_EXPORTER_LOG_FORMAT"); format != "" {
		baseConfig.Logging.Format = format
	} else {
		baseConfig.Logging.Format = "json"
	}

	// Metrics configuration
	baseConfig.Metrics.Collection.DefaultInterval = config.Duration{}
	baseConfig.Metrics.Collection.DefaultIntervalSet = false

	cfg.BaseConfig = *baseConfig

	// WebSocket configuration
	if url := os.Getenv("Z2M_EXPORTER_WEBSOCKET_URL"); url != "" {
		cfg.WebSocket.URL = url
	} else {
		cfg.WebSocket.URL = "ws://localhost:8081/api"
	}

	return cfg
}
