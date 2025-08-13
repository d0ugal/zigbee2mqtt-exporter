package config

import (
	"os"
	"testing"
)

func TestLoadFromEnvironment(t *testing.T) {
	// Test default values
	cfg := LoadFromEnvironment()
	
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default server host to be '0.0.0.0', got '%s'", cfg.Server.Host)
	}
	
	if cfg.Server.Port != 8087 {
		t.Errorf("Expected default server port to be 8087, got %d", cfg.Server.Port)
	}
	
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected default log level to be 'info', got '%s'", cfg.Logging.Level)
	}
	
	if cfg.Logging.Format != "json" {
		t.Errorf("Expected default log format to be 'json', got '%s'", cfg.Logging.Format)
	}
	
	if cfg.WebSocket.URL != "ws://localhost:8081/api" {
		t.Errorf("Expected default WebSocket URL to be 'ws://localhost:8081/api', got '%s'", cfg.WebSocket.URL)
	}
}

func TestLoadFromEnvironmentWithCustomValues(t *testing.T) {
	// Set custom environment variables
	os.Setenv("Z2M_EXPORTER_SERVER_HOST", "127.0.0.1")
	os.Setenv("Z2M_EXPORTER_SERVER_PORT", "9090")
	os.Setenv("Z2M_EXPORTER_LOG_LEVEL", "debug")
	os.Setenv("Z2M_EXPORTER_LOG_FORMAT", "text")
	os.Setenv("Z2M_EXPORTER_WEBSOCKET_URL", "ws://test:8081/api")
	
	defer func() {
		os.Unsetenv("Z2M_EXPORTER_SERVER_HOST")
		os.Unsetenv("Z2M_EXPORTER_SERVER_PORT")
		os.Unsetenv("Z2M_EXPORTER_LOG_LEVEL")
		os.Unsetenv("Z2M_EXPORTER_LOG_FORMAT")
		os.Unsetenv("Z2M_EXPORTER_WEBSOCKET_URL")
	}()
	
	cfg := LoadFromEnvironment()
	
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected server host to be '127.0.0.1', got '%s'", cfg.Server.Host)
	}
	
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server port to be 9090, got %d", cfg.Server.Port)
	}
	
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected log level to be 'debug', got '%s'", cfg.Logging.Level)
	}
	
	if cfg.Logging.Format != "text" {
		t.Errorf("Expected log format to be 'text', got '%s'", cfg.Logging.Format)
	}
	
	if cfg.WebSocket.URL != "ws://test:8081/api" {
		t.Errorf("Expected WebSocket URL to be 'ws://test:8081/api', got '%s'", cfg.WebSocket.URL)
	}
}

func TestLoadFromEnvironmentWithInvalidPort(t *testing.T) {
	// Set invalid port
	os.Setenv("Z2M_EXPORTER_SERVER_PORT", "invalid")
	
	defer func() {
		os.Unsetenv("Z2M_EXPORTER_SERVER_PORT")
	}()
	
	cfg := LoadFromEnvironment()
	
	// Should fall back to default port
	if cfg.Server.Port != 8087 {
		t.Errorf("Expected server port to fall back to 8087, got %d", cfg.Server.Port)
	}
}
