package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/logging"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/d0ugal/promexporter/version"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/collectors"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
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

	// Configure logging using promexporter
	logging.Configure(&logging.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
	})

	// Initialize metrics registry using promexporter
	metricsRegistry := promexporter_metrics.NewRegistry("zigbee2mqtt_exporter_info")

	// Add custom metrics to the registry
	z2mRegistry := metrics.NewZ2MRegistry(metricsRegistry)

	// Create collector
	z2mCollector := collectors.NewZ2MCollector(cfg, z2mRegistry)

	// Create and run application using promexporter
	application := app.New("Zigbee2MQTT Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(metricsRegistry).
		WithCollector(z2mCollector).
		Build()

	if err := application.Run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}
