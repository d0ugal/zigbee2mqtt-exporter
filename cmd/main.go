package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/collectors"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/logging"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/server"
)

// hasEnvironmentVariables checks if any Z2M_EXPORTER_* environment variables are set
func hasEnvironmentVariables() bool {
	envVars := []string{
		"Z2M_EXPORTER_SERVER_HOST",
		"Z2M_EXPORTER_SERVER_PORT",
		"Z2M_EXPORTER_LOG_LEVEL",
		"Z2M_EXPORTER_LOG_FORMAT",
		"Z2M_EXPORTER_WEBSOCKET_URL",
	}

	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

func main() {
	// Load configuration from environment variables
	cfg := config.LoadFromEnvironment()

	// Configure logging
	logging.Configure(&cfg.Logging)

	// Initialize metrics
	metricsRegistry := metrics.NewRegistry()

	// Create collectors
	z2mCollector := collectors.NewZ2MCollector(cfg, metricsRegistry)

	// Start collectors
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	z2mCollector.Start(ctx)

	// Create and start server
	srv := server.New(cfg, metricsRegistry)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Shutting down gracefully...")
		cancel()

		if err := srv.Shutdown(); err != nil {
			slog.Error("Failed to shutdown server gracefully", "error", err)
		}
	}()

	// Start server
	if err := srv.Start(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
