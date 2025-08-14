package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/collectors"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/logging"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/server"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/version"
)

func main() {
	// Parse command line flags
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version information")
	flag.Parse()

	// Show version if requested
	if showVersion {
		versionInfo := version.Get()
		fmt.Printf("zigbee2mqtt-exporter %s\n", versionInfo.Version)
		fmt.Printf("Commit: %s\n", versionInfo.Commit)
		fmt.Printf("Build Date: %s\n", versionInfo.BuildDate)
		fmt.Printf("Go Version: %s\n", versionInfo.GoVersion)
		os.Exit(0)
	}

	// Load configuration from environment variables
	cfg := config.LoadFromEnvironment()

	// Configure logging
	logging.Configure(&cfg.Logging)

	// Initialize metrics
	metricsRegistry := metrics.NewRegistry()

	// Set version info metric
	versionInfo := version.Get()
	metricsRegistry.VersionInfo.WithLabelValues(versionInfo.Version, versionInfo.Commit, versionInfo.BuildDate).Set(1)

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
