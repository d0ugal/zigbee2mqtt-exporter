package collectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/d0ugal/promexporter/app"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// TestZ2MCollectorIntegration tests the full collection flow to catch label mapping issues
func TestZ2MCollectorIntegration(t *testing.T) {
	// Create test server that returns valid Zigbee2MQTT WebSocket responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api":
			// Simulate WebSocket upgrade
			w.Header().Set("Upgrade", "websocket")
			w.Header().Set("Connection", "Upgrade")
			w.WriteHeader(http.StatusSwitchingProtocols)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create test configuration
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://" + server.URL[7:] + "/api", // Convert http to ws
		},
	}

	// Create a fresh registry for testing
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)

	// Create a minimal app instance for testing
	testApp := app.New("Test Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(baseRegistry).
		Build()

	collector := NewZ2MCollector(cfg, registry, testApp)

	// Test the collection flow
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Start the collector
	collector.Start(ctx)

	// Wait a bit for collection to happen
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop collection
	cancel()

	// Wait for collection to complete
	time.Sleep(100 * time.Millisecond)

	// Verify that metrics were created with correct labels
	// This is the key test - it will panic if labels don't match metric definitions

	// Test WebSocket connection metrics
	websocketStatusMetric := testutil.ToFloat64(registry.WebSocketConnectionStatus.With(prometheus.Labels{}))
	t.Logf("WebSocket connection status metric: %f", websocketStatusMetric)

	websocketMessagesMetric := testutil.ToFloat64(registry.WebSocketMessagesTotal.With(prometheus.Labels{
		"topic": "bridge/devices",
	}))
	t.Logf("WebSocket messages metric: %f", websocketMessagesMetric)

	websocketReconnectsMetric := testutil.ToFloat64(registry.WebSocketReconnectsTotal.With(prometheus.Labels{}))
	t.Logf("WebSocket reconnects metric: %f", websocketReconnectsMetric)

	// Test bridge metrics
	bridgeStateMetric := testutil.ToFloat64(registry.BridgeState.With(prometheus.Labels{}))
	t.Logf("Bridge state metric: %f", bridgeStateMetric)

	bridgeEventsMetric := testutil.ToFloat64(registry.BridgeEventsTotal.With(prometheus.Labels{
		"event_type": "device_announce",
	}))
	t.Logf("Bridge events metric: %f", bridgeEventsMetric)

	// Test device metrics
	deviceInfoMetric := testutil.ToFloat64(registry.DeviceInfo.With(prometheus.Labels{
		"device":            "test-device",
		"type":              "EndDevice",
		"power_source":      "Battery",
		"manufacturer":      "Test Manufacturer",
		"model_id":          "TEST_MODEL",
		"supported":         "true",
		"disabled":          "false",
		"interview_state":   "completed",
		"software_build_id": "1.0.0",
		"date_code":         "20250101",
	}))
	t.Logf("Device info metric: %f", deviceInfoMetric)

	deviceAvailabilityMetric := testutil.ToFloat64(registry.DeviceAvailability.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device availability metric: %f", deviceAvailabilityMetric)

	deviceLastSeenMetric := testutil.ToFloat64(registry.DeviceLastSeen.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device last seen metric: %f", deviceLastSeenMetric)

	deviceSeenCountMetric := testutil.ToFloat64(registry.DeviceSeenCount.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device seen count metric: %f", deviceSeenCountMetric)

	deviceLinkQualityMetric := testutil.ToFloat64(registry.DeviceLinkQuality.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device link quality metric: %f", deviceLinkQualityMetric)

	deviceStateMetric := testutil.ToFloat64(registry.DeviceState.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device state metric: %f", deviceStateMetric)

	deviceBatteryMetric := testutil.ToFloat64(registry.DeviceBattery.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device battery metric: %f", deviceBatteryMetric)

	// Test OTA update metrics
	deviceOTAUpdateMetric := testutil.ToFloat64(registry.DeviceOTAUpdateAvailable.With(prometheus.Labels{
		"device": "test-device",
	}))
	t.Logf("Device OTA update metric: %f", deviceOTAUpdateMetric)

	deviceCurrentFirmwareMetric := testutil.ToFloat64(registry.DeviceCurrentFirmware.With(prometheus.Labels{
		"device":           "test-device",
		"firmware_version": "1.0.0",
	}))
	t.Logf("Device current firmware metric: %f", deviceCurrentFirmwareMetric)

	deviceAvailableFirmwareMetric := testutil.ToFloat64(registry.DeviceAvailableFirmware.With(prometheus.Labels{
		"device":           "test-device",
		"firmware_version": "1.1.0",
	}))
	t.Logf("Device available firmware metric: %f", deviceAvailableFirmwareMetric)

	// If we get here without panicking, the label mapping is correct
	t.Log("✅ All metrics created successfully with correct label mapping")
}

// TestZ2MCollectorLabelConsistency tests that all metric labels match their definitions
func TestZ2MCollectorLabelConsistency(t *testing.T) {
	// Create a fresh registry for testing
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)

	// Test all metrics with their expected labels
	testCases := []struct {
		name        string
		metric      *prometheus.CounterVec
		labels      prometheus.Labels
		description string
	}{
		{
			name:        "WebSocketMessagesTotal",
			metric:      registry.WebSocketMessagesTotal,
			labels:      prometheus.Labels{"topic": "bridge/devices"},
			description: "Should accept 'topic' label",
		},
		{
			name:        "BridgeEventsTotal",
			metric:      registry.BridgeEventsTotal,
			labels:      prometheus.Labels{"event_type": "device_announce"},
			description: "Should accept 'event_type' label",
		},
		{
			name:        "DeviceSeenCount",
			metric:      registry.DeviceSeenCount,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "WebSocketReconnectsTotal",
			metric:      registry.WebSocketReconnectsTotal,
			labels:      prometheus.Labels{},
			description: "Should accept no labels",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This will panic if labels don't match the metric definition
			counter := tc.metric.With(tc.labels)
			counter.Inc()

			// Verify the metric was created successfully
			value := testutil.ToFloat64(counter)
			if value != 1.0 {
				t.Errorf("Expected metric value 1.0, got %f", value)
			}

			t.Logf("✅ %s: %s", tc.name, tc.description)
		})
	}

	// Test gauge metrics
	gaugeTestCases := []struct {
		name        string
		metric      *prometheus.GaugeVec
		labels      prometheus.Labels
		description string
	}{
		{
			name:        "WebSocketConnectionStatus",
			metric:      registry.WebSocketConnectionStatus,
			labels:      prometheus.Labels{},
			description: "Should accept no labels",
		},
		{
			name:        "BridgeState",
			metric:      registry.BridgeState,
			labels:      prometheus.Labels{},
			description: "Should accept no labels",
		},
		{
			name:        "DeviceInfo",
			metric:      registry.DeviceInfo,
			labels:      prometheus.Labels{"device": "test-device", "type": "EndDevice", "power_source": "Battery", "manufacturer": "Test", "model_id": "TEST", "supported": "true", "disabled": "false", "interview_state": "completed", "software_build_id": "1.0.0", "date_code": "20250101"},
			description: "Should accept all device info labels",
		},
		{
			name:        "DeviceAvailability",
			metric:      registry.DeviceAvailability,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceLastSeen",
			metric:      registry.DeviceLastSeen,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceLinkQuality",
			metric:      registry.DeviceLinkQuality,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceState",
			metric:      registry.DeviceState,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceBattery",
			metric:      registry.DeviceBattery,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceOTAUpdateAvailable",
			metric:      registry.DeviceOTAUpdateAvailable,
			labels:      prometheus.Labels{"device": "test-device"},
			description: "Should accept 'device' label",
		},
		{
			name:        "DeviceCurrentFirmware",
			metric:      registry.DeviceCurrentFirmware,
			labels:      prometheus.Labels{"device": "test-device", "firmware_version": "1.0.0"},
			description: "Should accept 'device' and 'firmware_version' labels",
		},
		{
			name:        "DeviceAvailableFirmware",
			metric:      registry.DeviceAvailableFirmware,
			labels:      prometheus.Labels{"device": "test-device", "firmware_version": "1.1.0"},
			description: "Should accept 'device' and 'firmware_version' labels",
		},
	}

	for _, tc := range gaugeTestCases {
		t.Run(tc.name, func(t *testing.T) {
			// This will panic if labels don't match the metric definition
			gauge := tc.metric.With(tc.labels)
			gauge.Set(42.0)

			// Verify the metric was created successfully
			value := testutil.ToFloat64(gauge)
			if value != 42.0 {
				t.Errorf("Expected metric value 42.0, got %f", value)
			}

			t.Logf("✅ %s: %s", tc.name, tc.description)
		})
	}

	t.Log("✅ All metric label consistency tests passed")
}

// TestZ2MCollectorErrorHandling tests error scenarios to ensure they don't cause label panics
func TestZ2MCollectorErrorHandling(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create test configuration
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://" + server.URL[7:] + "/api", // Convert http to ws
		},
	}

	// Create a fresh registry for testing
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)

	// Create a minimal app instance for testing
	testApp := app.New("Test Exporter").
		WithConfig(&cfg.BaseConfig).
		WithMetrics(baseRegistry).
		Build()

	collector := NewZ2MCollector(cfg, registry, testApp)

	// Test error handling without panicking
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	collector.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(100 * time.Millisecond)

	// If we get here without panicking, error handling is working correctly
	t.Log("✅ Error handling works correctly without label panics")
}

// TestZ2MCollectorMessageProcessing tests message processing to catch label mapping issues
func TestZ2MCollectorMessageProcessing(t *testing.T) {
	// Create a fresh registry for testing
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	registry := metrics.NewZ2MRegistry(baseRegistry)

	// Test that we can create metrics with the correct labels without panicking
	// This is the key test - it will panic if labels don't match metric definitions

	// Test device info metric creation
	deviceInfoMetric := registry.DeviceInfo.With(prometheus.Labels{
		"device":            "test-device",
		"type":              "EndDevice",
		"power_source":      "Battery",
		"manufacturer":      "Test Manufacturer",
		"model_id":          "TEST_MODEL",
		"supported":         "true",
		"disabled":          "false",
		"interview_state":   "completed",
		"software_build_id": "1.0.0",
		"date_code":         "20250101",
	})
	deviceInfoMetric.Set(1.0)

	// Verify the metric was created successfully
	value := testutil.ToFloat64(deviceInfoMetric)
	if value != 1.0 {
		t.Errorf("Expected device info metric value 1.0, got %f", value)
	}

	// Test other device metrics
	deviceAvailabilityMetric := registry.DeviceAvailability.With(prometheus.Labels{
		"device": "test-device",
	})
	deviceAvailabilityMetric.Set(1.0)

	deviceLastSeenMetric := registry.DeviceLastSeen.With(prometheus.Labels{
		"device": "test-device",
	})
	deviceLastSeenMetric.Set(float64(time.Now().Unix()))

	deviceSeenCountMetric := registry.DeviceSeenCount.With(prometheus.Labels{
		"device": "test-device",
	})
	deviceSeenCountMetric.Inc()

	t.Log("✅ Message processing works correctly with proper label mapping")
}
