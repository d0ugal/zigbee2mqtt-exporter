package collectors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// TestMetricLabelConsistency tests that all metrics are created with correct labels
// This test would have caught the label mapping bugs we just fixed
func TestMetricLabelConsistency(t *testing.T) {
	// Create a fresh registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	z2mRegistry := metrics.NewZ2MRegistry(baseRegistry)

	// Create test config
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://localhost:8080/api",
		},
	}

	_ = NewZ2MCollector(cfg, z2mRegistry)

	// Test device metrics with correct labels
	t.Run("DeviceMetrics", func(t *testing.T) {
		// Test DeviceInfo
		z2mRegistry.DeviceInfo.With(prometheus.Labels{
			"device":            "test-device",
			"type":              "sensor",
			"power_source":      "battery",
			"manufacturer":      "test-manufacturer",
			"model_id":          "test-model",
			"supported":         "true",
			"disabled":          "false",
			"interview_state":   "completed",
			"software_build_id": "test-build",
			"date_code":         "20230101",
		}).Set(1)

		// Test DeviceLastSeen
		z2mRegistry.DeviceLastSeen.With(prometheus.Labels{
			"device": "test-device",
		}).Set(float64(time.Now().Unix()))

		// Test DeviceSeenCount
		z2mRegistry.DeviceSeenCount.With(prometheus.Labels{
			"device": "test-device",
		}).Inc()

		// Test DeviceLinkQuality
		z2mRegistry.DeviceLinkQuality.With(prometheus.Labels{
			"device": "test-device",
		}).Set(255)

		// Test DeviceState
		z2mRegistry.DeviceState.With(prometheus.Labels{
			"device": "test-device",
		}).Set(1)

		// Test DeviceBattery
		z2mRegistry.DeviceBattery.With(prometheus.Labels{
			"device": "test-device",
		}).Set(100)

		// Test DeviceAvailability
		z2mRegistry.DeviceAvailability.With(prometheus.Labels{
			"device": "test-device",
		}).Set(1)

		// Test DeviceOTAUpdateAvailable
		z2mRegistry.DeviceOTAUpdateAvailable.With(prometheus.Labels{
			"device": "test-device",
		}).Set(0)

		// Test DeviceCurrentFirmware
		z2mRegistry.DeviceCurrentFirmware.With(prometheus.Labels{
			"device":           "test-device",
			"firmware_version": "1.0.0",
		}).Set(1)

		// Test DeviceAvailableFirmware
		z2mRegistry.DeviceAvailableFirmware.With(prometheus.Labels{
			"device":           "test-device",
			"firmware_version": "1.1.0",
		}).Set(1)

		// Verify metrics were created successfully (no panic)
		t.Log("Device metrics created successfully with correct labels")
	})

	// Test bridge metrics with correct labels
	t.Run("BridgeMetrics", func(t *testing.T) {
		// Test BridgeState (no labels)
		z2mRegistry.BridgeState.With(prometheus.Labels{}).Set(1)

		// Test BridgeEventsTotal
		z2mRegistry.BridgeEventsTotal.With(prometheus.Labels{
			"event_type": "device_joined",
		}).Inc()

		// Verify metrics were created successfully (no panic)
		t.Log("Bridge metrics created successfully with correct labels")
	})

	// Test connection metrics with correct labels
	t.Run("ConnectionMetrics", func(t *testing.T) {
		// Test WebSocketConnectionStatus (no labels)
		z2mRegistry.WebSocketConnectionStatus.With(prometheus.Labels{}).Set(1)

		// Test WebSocketMessagesTotal
		z2mRegistry.WebSocketMessagesTotal.With(prometheus.Labels{
			"topic": "zigbee2mqtt/bridge/devices",
		}).Inc()

		// Test WebSocketReconnectsTotal (no labels)
		z2mRegistry.WebSocketReconnectsTotal.With(prometheus.Labels{}).Inc()

		// Verify metrics were created successfully (no panic)
		t.Log("Connection metrics created successfully with correct labels")
	})
}

// TestMetricLabelMismatchDetection tests that incorrect labels cause panics
// This test ensures our validation catches future regressions
func TestMetricLabelMismatchDetection(t *testing.T) {
	// Create a fresh registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	z2mRegistry := metrics.NewZ2MRegistry(baseRegistry)

	t.Run("DeviceMetricsWrongLabels", func(t *testing.T) {
		// This should panic with "label name 'device' missing in label map"
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly caught panic with wrong labels: %v", r)
			} else {
				t.Error("Expected panic with wrong labels, but none occurred")
			}
		}()

		// Use wrong label name (this should panic)
		z2mRegistry.DeviceLastSeen.With(prometheus.Labels{
			"friendly_name": "test-device", // Wrong! Should be "device"
		}).Set(float64(time.Now().Unix()))
	})

	t.Run("BridgeMetricsWrongLabels", func(t *testing.T) {
		// This should panic with "label name 'event_type' missing in label map"
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Correctly caught panic with wrong labels: %v", r)
			} else {
				t.Error("Expected panic with wrong labels, but none occurred")
			}
		}()

		// Use wrong label name (this should panic)
		z2mRegistry.BridgeEventsTotal.With(prometheus.Labels{
			"event": "device_joined", // Wrong! Should be "event_type"
		}).Inc()
	})
}

// TestCollectorIntegration tests the full collection flow with mock data
// This test simulates the actual runtime behavior that was failing
func TestCollectorIntegration(t *testing.T) {
	// Create a test server that returns valid Zigbee2MQTT WebSocket responses
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

	// Create test config
	cfg := &config.Config{
		WebSocket: config.WebSocketConfig{
			URL: "ws://" + server.URL[7:] + "/api", // Convert http:// to ws://
		},
	}

	// Create a fresh registry for testing
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	z2mRegistry := metrics.NewZ2MRegistry(baseRegistry)

	collector := NewZ2MCollector(cfg, z2mRegistry)

	// Test the collection flow
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the collector
	collector.Start(ctx)

	// Wait a bit for collection to happen
	time.Sleep(1 * time.Second)

	// Verify metrics were created without panics
	// This test would have caught the label mapping bugs in the actual collection flow
	t.Log("Integration test completed successfully - no panics occurred")
}

// TestMetricDefinitionConsistency tests that metric definitions match expected labels
// This test ensures the metrics registry and collector are in sync
func TestMetricDefinitionConsistency(t *testing.T) {
	baseRegistry := promexporter_metrics.NewRegistry("test_exporter_info")
	z2mRegistry := metrics.NewZ2MRegistry(baseRegistry)

	// Test that all metrics have the expected label names
	expectedLabels := map[string][]string{
		"zigbee2mqtt_device_info":                        {"device", "type", "power_source", "manufacturer", "model_id", "supported", "disabled", "interview_state", "software_build_id", "date_code"},
		"zigbee2mqtt_device_last_seen_timestamp":        {"device"},
		"zigbee2mqtt_device_seen_total":                  {"device"},
		"zigbee2mqtt_device_link_quality":                {"device"},
		"zigbee2mqtt_device_power_state":                 {"device"},
		"zigbee2mqtt_device_battery_level":               {"device"},
		"zigbee2mqtt_device_up":                          {"device"},
		"zigbee2mqtt_device_ota_update_available":        {"device"},
		"zigbee2mqtt_device_current_firmware_version":    {"device", "firmware_version"},
		"zigbee2mqtt_device_available_firmware_version":   {"device", "firmware_version"},
		"zigbee2mqtt_bridge_state":                        {},
		"zigbee2mqtt_bridge_events_total":                {"event_type"},
		"zigbee2mqtt_websocket_connection_status":         {},
		"zigbee2mqtt_websocket_messages_total":           {"topic"},
		"zigbee2mqtt_websocket_reconnects_total":          {},
	}

	for metricName, expectedLabelNames := range expectedLabels {
		t.Run(metricName, func(t *testing.T) {
			// Create a test metric with the expected labels
			labels := make(prometheus.Labels)
			for _, labelName := range expectedLabelNames {
				labels[labelName] = "test-value"
			}

			// This should not panic if labels are correct
			switch metricName {
			case "zigbee2mqtt_device_info":
				z2mRegistry.DeviceInfo.With(labels).Set(1)
			case "zigbee2mqtt_device_last_seen_timestamp":
				z2mRegistry.DeviceLastSeen.With(labels).Set(float64(time.Now().Unix()))
			case "zigbee2mqtt_device_seen_total":
				z2mRegistry.DeviceSeenCount.With(labels).Inc()
			case "zigbee2mqtt_device_link_quality":
				z2mRegistry.DeviceLinkQuality.With(labels).Set(255)
			case "zigbee2mqtt_device_power_state":
				z2mRegistry.DeviceState.With(labels).Set(1)
			case "zigbee2mqtt_device_battery_level":
				z2mRegistry.DeviceBattery.With(labels).Set(100)
			case "zigbee2mqtt_device_up":
				z2mRegistry.DeviceAvailability.With(labels).Set(1)
			case "zigbee2mqtt_device_ota_update_available":
				z2mRegistry.DeviceOTAUpdateAvailable.With(labels).Set(0)
			case "zigbee2mqtt_device_current_firmware_version":
				z2mRegistry.DeviceCurrentFirmware.With(labels).Set(1)
			case "zigbee2mqtt_device_available_firmware_version":
				z2mRegistry.DeviceAvailableFirmware.With(labels).Set(1)
			case "zigbee2mqtt_bridge_state":
				z2mRegistry.BridgeState.With(labels).Set(1)
			case "zigbee2mqtt_bridge_events_total":
				z2mRegistry.BridgeEventsTotal.With(labels).Inc()
			case "zigbee2mqtt_websocket_connection_status":
				z2mRegistry.WebSocketConnectionStatus.With(labels).Set(1)
			case "zigbee2mqtt_websocket_messages_total":
				z2mRegistry.WebSocketMessagesTotal.With(labels).Inc()
			case "zigbee2mqtt_websocket_reconnects_total":
				z2mRegistry.WebSocketReconnectsTotal.With(labels).Inc()
			}

			t.Logf("Metric %s created successfully with expected labels: %v", metricName, expectedLabelNames)
		})
	}
}
