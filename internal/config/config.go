package config

import (
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Logging   LoggingConfig   `yaml:"logging"`
	WebSocket WebSocketConfig `yaml:"websocket"`
}

// ServerConfig represents the HTTP server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LoggingConfig represents the logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// WebSocketConfig represents the WebSocket configuration
type WebSocketConfig struct {
	URL string `yaml:"url"`
}

// LoadFromEnvironment loads configuration from environment variables
func LoadFromEnvironment() *Config {
	cfg := &Config{}

	// Server configuration
	if host := os.Getenv("Z2M_EXPORTER_SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}

	if portStr := os.Getenv("Z2M_EXPORTER_SERVER_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// Logging configuration
	if level := os.Getenv("Z2M_EXPORTER_LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}

	if format := os.Getenv("Z2M_EXPORTER_LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}

	// WebSocket configuration
	if url := os.Getenv("Z2M_EXPORTER_WEBSOCKET_URL"); url != "" {
		cfg.WebSocket.URL = url
	}

	// Set defaults
	setDefaults(cfg)

	return cfg
}

// setDefaults sets default values for configuration
func setDefaults(cfg *Config) {
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}

	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8087
	}

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}

	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}

	if cfg.WebSocket.URL == "" {
		cfg.WebSocket.URL = "ws://localhost:8081/api"
	}
}
