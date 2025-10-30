package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/d0ugal/promexporter/app"
	"github.com/d0ugal/promexporter/tracing"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/config"
	"github.com/d0ugal/zigbee2mqtt-exporter/internal/metrics"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

// Z2MCollector handles WebSocket connections to Zigbee2MQTT
type Z2MCollector struct {
	cfg     *config.Config
	metrics *metrics.Z2MRegistry
	app     *app.App
	conn    *websocket.Conn
	done    chan struct{}
	// Device metadata cache - maps device name to device info
	deviceInfo map[string]DeviceInfo
}

// DeviceInfo stores device metadata from bridge/devices message
type DeviceInfo struct {
	Type            string
	PowerSource     string
	Manufacturer    string
	ModelID         string
	Supported       bool
	Disabled        bool
	InterviewState  string
	NetworkAddress  int
	SoftwareBuildID string
	DateCode        string
}

// Z2MMessage represents a message from Zigbee2MQTT
type Z2MMessage struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

// BridgeDevice represents a device from bridge/devices message
type BridgeDevice struct {
	IEEEAddress        string `json:"ieee_address"`
	Type               string `json:"type"`
	NetworkAddress     int    `json:"network_address"`
	Supported          bool   `json:"supported"`
	FriendlyName       string `json:"friendly_name"`
	Disabled           bool   `json:"disabled"`
	Description        string `json:"description"`
	PowerSource        string `json:"power_source"`
	SoftwareBuildID    string `json:"software_build_id"`
	DateCode           string `json:"date_code"`
	ModelID            string `json:"model_id"`
	Interviewing       bool   `json:"interviewing"`
	InterviewCompleted bool   `json:"interview_completed"`
	InterviewState     string `json:"interview_state"`
	Manufacturer       string `json:"manufacturer"`
	// OTA-related fields (may be available depending on device)
	AvailableFirmwareVersion string `json:"available_firmware_version,omitempty"`
	CurrentFirmwareVersion   string `json:"current_firmware_version,omitempty"`
	UpdateAvailable          bool   `json:"update_available,omitempty"`
}

// NewZ2MCollector creates a new Z2M collector
func NewZ2MCollector(cfg *config.Config, metrics *metrics.Z2MRegistry, app *app.App) *Z2MCollector {
	return &Z2MCollector{
		cfg:        cfg,
		metrics:    metrics,
		app:        app,
		done:       make(chan struct{}),
		deviceInfo: make(map[string]DeviceInfo),
	}
}

// Start begins collecting metrics from Zigbee2MQTT
func (c *Z2MCollector) Start(ctx context.Context) {
	go c.run(ctx)
}

// run handles the main collection loop with automatic reconnection
func (c *Z2MCollector) run(ctx context.Context) {
	reconnectDelay := time.Second
	maxReconnectDelay := time.Minute

	for {
		// Create a span for each connection attempt
		tracer := c.app.GetTracer()

		var collectorSpan *tracing.CollectorSpan

		if tracer != nil && tracer.IsEnabled() {
			collectorSpan = tracer.NewCollectorSpan(ctx, "z2m-collector", "connection-attempt")
			collectorSpan.SetAttributes(
				attribute.String("websocket.url", c.cfg.WebSocket.URL),
			)
		}

		select {
		case <-ctx.Done():
			if collectorSpan != nil {
				collectorSpan.AddEvent("shutdown_requested")
				collectorSpan.End()
			}

			return
		case <-c.done:
			if collectorSpan != nil {
				collectorSpan.AddEvent("stop_requested")
				collectorSpan.End()
			}

			return
		default:
		}

		if err := c.connect(); err != nil {
			slog.Error("Failed to connect to Zigbee2MQTT",
				"error", err,
				"url", c.cfg.WebSocket.URL,
				"reconnect_delay", reconnectDelay,
			)

			if collectorSpan != nil {
				collectorSpan.RecordError(err, attribute.String("websocket.url", c.cfg.WebSocket.URL))
				collectorSpan.End()
			}

			c.metrics.WebSocketConnectionStatus.With(prometheus.Labels{}).Set(0)
			c.metrics.WebSocketReconnectsTotal.With(prometheus.Labels{}).Inc()

			select {
			case <-ctx.Done():
				return
			case <-time.After(reconnectDelay):
				slog.Info("Attempting Zigbee2MQTT reconnection",
					"url", c.cfg.WebSocket.URL,
					"delay", reconnectDelay,
				)
				reconnectDelay = minDuration(reconnectDelay*2, maxReconnectDelay)

				continue
			}
		}

		// Reset reconnect delay on successful connection
		reconnectDelay = time.Second

		if collectorSpan != nil {
			collectorSpan.AddEvent("connection_successful")
			collectorSpan.End()
		}

		slog.Info("Successfully connected to Zigbee2MQTT", "url", c.cfg.WebSocket.URL)
		c.metrics.WebSocketConnectionStatus.With(prometheus.Labels{}).Set(1)

		// Start reading messages
		if err := c.readMessages(ctx); err != nil {
			slog.Error("Error reading messages", "error", err)
			c.metrics.WebSocketConnectionStatus.With(prometheus.Labels{}).Set(0)
		}
	}
}

// connect establishes a WebSocket connection to Zigbee2MQTT
func (c *Z2MCollector) connect() error {
	slog.Info("Connecting to Zigbee2MQTT", "url", c.cfg.WebSocket.URL)

	conn, _, err := websocket.DefaultDialer.Dial(c.cfg.WebSocket.URL, nil)
	if err != nil {
		return fmt.Errorf("failed to dial WebSocket: %w", err)
	}

	c.conn = conn

	slog.Info("Connected to Zigbee2MQTT WebSocket")

	return nil
}

// readMessages reads and processes messages from the WebSocket
func (c *Z2MCollector) readMessages(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				return fmt.Errorf("unexpected close error: %w", err)
			}

			return fmt.Errorf("read message error: %w", err)
		}

		c.processMessage(ctx, message)
	}
}

// processMessage processes a single WebSocket message
func (c *Z2MCollector) processMessage(ctx context.Context, message []byte) {
	// Create span for message processing
	tracer := c.app.GetTracer()

	var collectorSpan *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		collectorSpan = tracer.NewCollectorSpan(ctx, "z2m-collector", "process-message")
		defer collectorSpan.End()
	}

	var z2mMsg Z2MMessage
	if err := json.Unmarshal(message, &z2mMsg); err != nil {
		slog.Error("Failed to unmarshal message", "error", err, "message", string(message))

		if collectorSpan != nil {
			collectorSpan.RecordError(err)
		}

		return
	}

	if collectorSpan != nil {
		collectorSpan.SetAttributes(
			attribute.String("message.topic", z2mMsg.Topic),
			attribute.Int("message.payload_length", len(message)),
		)
	}

	// Increment message counter
	c.metrics.WebSocketMessagesTotal.With(prometheus.Labels{
		"topic": z2mMsg.Topic,
	}).Inc()

	// Process different message types
	switch z2mMsg.Topic {
	case "bridge/logging":
		c.processLoggingMessage(z2mMsg)
	case "bridge/devices":
		c.processBridgeDevicesMessage(z2mMsg)
	case "bridge/state":
		c.processBridgeStateMessage(z2mMsg)
	case "bridge/event":
		c.processBridgeEventMessage(z2mMsg)
	default:
		// This is likely a device message
		if z2mMsg.Topic != "" && !strings.HasPrefix(z2mMsg.Topic, "bridge/") {
			// Check if this is an availability message
			if strings.HasSuffix(z2mMsg.Topic, "/availability") {
				c.processAvailabilityMessage(z2mMsg)
			} else {
				c.processDeviceMessage(z2mMsg)
			}
		}
	}

	if collectorSpan != nil {
		collectorSpan.AddEvent("message_processed",
			attribute.String("message.topic", z2mMsg.Topic),
		)
	}
}

// processBridgeDevicesMessage processes bridge/devices messages to cache device types
func (c *Z2MCollector) processBridgeDevicesMessage(msg Z2MMessage) {
	devicesData, ok := msg.Payload.([]interface{})
	if !ok {
		// Try parsing as array directly
		var devices []BridgeDevice
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", msg.Payload)), &devices); err != nil {
			slog.Debug("Could not parse bridge/devices message", "error", err)
			return
		}

		// Cache device info and update device info metric
		for _, device := range devices {
			powerSource := device.PowerSource
			if powerSource == "" {
				powerSource = "unknown"
			}

			// Convert boolean values to strings for labels
			supported := "false"
			if device.Supported {
				supported = "true"
			}

			disabled := "false"
			if device.Disabled {
				disabled = "true"
			}

			// Map interview state to string
			interviewState := "not_started"

			switch device.InterviewState {
			case "in_progress":
				interviewState = "in_progress"
			case "successful":
				interviewState = "successful"
			}

			// Handle firmware information
			softwareBuildID := device.SoftwareBuildID
			if softwareBuildID == "" {
				softwareBuildID = "unknown"
			}

			dateCode := device.DateCode
			if dateCode == "" {
				dateCode = "unknown"
			}

			c.deviceInfo[device.FriendlyName] = DeviceInfo{
				Type:            device.Type,
				PowerSource:     powerSource,
				Manufacturer:    device.Manufacturer,
				ModelID:         device.ModelID,
				Supported:       device.Supported,
				Disabled:        device.Disabled,
				InterviewState:  device.InterviewState,
				NetworkAddress:  device.NetworkAddress,
				SoftwareBuildID: softwareBuildID,
				DateCode:        dateCode,
			}

			// Update device info metric (always set to 1)
			c.metrics.DeviceInfo.With(prometheus.Labels{
				"device":            device.FriendlyName,
				"type":              device.Type,
				"power_source":      powerSource,
				"manufacturer":      device.Manufacturer,
				"model_id":          device.ModelID,
				"supported":         supported,
				"disabled":          disabled,
				"interview_state":   interviewState,
				"software_build_id": softwareBuildID,
				"date_code":         dateCode,
			}).Set(1)

			// Update OTA metrics
			if device.CurrentFirmwareVersion != "" {
				c.metrics.DeviceCurrentFirmware.With(prometheus.Labels{
					"device":  device.FriendlyName,
					"version": device.CurrentFirmwareVersion,
				}).Set(1)
			}

			if device.AvailableFirmwareVersion != "" {
				c.metrics.DeviceAvailableFirmware.With(prometheus.Labels{
					"device":  device.FriendlyName,
					"version": device.AvailableFirmwareVersion,
				}).Set(1)
			}

			// Update OTA update availability
			updateAvailable := 0.0
			if device.UpdateAvailable {
				updateAvailable = 1.0
			}

			c.metrics.DeviceOTAUpdateAvailable.With(prometheus.Labels{
				"device": device.FriendlyName,
			}).Set(updateAvailable)
		}

		return
	}

	// Parse devices array
	for _, deviceData := range devicesData {
		deviceBytes, err := json.Marshal(deviceData)
		if err != nil {
			continue
		}

		var device BridgeDevice
		if err := json.Unmarshal(deviceBytes, &device); err != nil {
			continue
		}

		powerSource := device.PowerSource
		if powerSource == "" {
			powerSource = "unknown"
		}

		// Convert boolean values to strings for labels
		supported := "false"
		if device.Supported {
			supported = "true"
		}

		disabled := "false"
		if device.Disabled {
			disabled = "true"
		}

		// Map interview state to string
		interviewState := "not_started"

		switch device.InterviewState {
		case "in_progress":
			interviewState = "in_progress"
		case "successful":
			interviewState = "successful"
		}

		c.deviceInfo[device.FriendlyName] = DeviceInfo{
			Type:            device.Type,
			PowerSource:     powerSource,
			Manufacturer:    device.Manufacturer,
			ModelID:         device.ModelID,
			Supported:       device.Supported,
			Disabled:        device.Disabled,
			InterviewState:  device.InterviewState,
			NetworkAddress:  device.NetworkAddress,
			SoftwareBuildID: device.SoftwareBuildID,
			DateCode:        device.DateCode,
		}

		// Update device info metric (always set to 1)
		c.metrics.DeviceInfo.With(prometheus.Labels{
			"device":            device.FriendlyName,
			"type":              device.Type,
			"power_source":      powerSource,
			"manufacturer":      device.Manufacturer,
			"model_id":          device.ModelID,
			"supported":         supported,
			"disabled":          disabled,
			"interview_state":   interviewState,
			"software_build_id": device.SoftwareBuildID,
			"date_code":         device.DateCode,
		}).Set(1)
	}

	slog.Debug("Updated device info cache", "device_count", len(c.deviceInfo))
}

// processBridgeStateMessage processes bridge/state messages
func (c *Z2MCollector) processBridgeStateMessage(msg Z2MMessage) {
	payloadMap, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	state, ok := payloadMap["state"].(string)
	if !ok {
		return
	}

	// Update bridge state metric
	if state == "online" {
		c.metrics.BridgeState.With(prometheus.Labels{}).Set(1)
	} else {
		c.metrics.BridgeState.With(prometheus.Labels{}).Set(0)
	}
}

// processBridgeEventMessage processes bridge/event messages
func (c *Z2MCollector) processBridgeEventMessage(msg Z2MMessage) {
	payloadMap, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	eventType, ok := payloadMap["type"].(string)
	if !ok {
		return
	}

	c.metrics.BridgeEventsTotal.With(prometheus.Labels{
		"event_type": eventType,
	}).Inc()
}

// processLoggingMessage processes bridge logging messages
func (c *Z2MCollector) processLoggingMessage(msg Z2MMessage) {
	payloadMap, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return
	}

	payload, ok := payloadMap["message"].(string)
	if !ok {
		return
	}

	// Look for MQTT publish messages that contain device data
	if strings.Contains(payload, "z2m:mqtt: MQTT publish:") {
		c.extractDeviceDataFromLogging(payload)
	}
}

// extractDeviceDataFromLogging extracts device data from logging messages
func (c *Z2MCollector) extractDeviceDataFromLogging(message string) {
	// Example: "z2m:mqtt: MQTT publish: topic 'zigbee2mqtt/Kitchen Air', payload '{\"co2\":360,\"formaldehyd\":2,\"humidity\":53.7,\"last_seen\":\"2025-08-12T15:16:05.916Z\",\"linkquality\":76,\"temperature\":25.1,\"voc\":8}'"

	// Extract topic
	topicStart := strings.Index(message, "topic '") + 7

	topicEnd := strings.Index(message[topicStart:], "'")
	if topicStart < 7 || topicEnd < 0 {
		return
	}

	deviceTopic := message[topicStart : topicStart+topicEnd]
	deviceName := strings.TrimPrefix(deviceTopic, "zigbee2mqtt/")

	// Skip availability messages (they're handled separately)
	if strings.HasSuffix(deviceName, "/availability") {
		return
	}

	// Extract payload
	payloadStart := strings.Index(message, "payload '") + 9

	payloadEnd := strings.LastIndex(message, "'")
	if payloadStart < 9 || payloadEnd < payloadStart {
		return
	}

	payloadStr := message[payloadStart:payloadEnd]

	// Parse device data
	var deviceData map[string]interface{}
	if err := json.Unmarshal([]byte(payloadStr), &deviceData); err != nil {
		slog.Debug("Could not parse device data from logging message", "error", err)
		return
	}

	c.updateDeviceMetrics(deviceName, deviceData)
}

// processAvailabilityMessage processes device availability messages
func (c *Z2MCollector) processAvailabilityMessage(msg Z2MMessage) {
	// Extract device name from availability topic (remove zigbee2mqtt/ prefix and /availability suffix)
	deviceName := strings.TrimPrefix(msg.Topic, "zigbee2mqtt/")
	deviceName = strings.TrimSuffix(deviceName, "/availability")

	// Handle availability payload - try different payload formats
	availabilityValue := 0.0

	// Try string payload first (e.g., "online" or "offline")
	if availability, ok := msg.Payload.(string); ok {
		if availability == "online" {
			availabilityValue = 1.0
		}
	} else if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
		// Try map payload with "state" field
		if state, ok := payloadMap["state"].(string); ok {
			if state == "online" {
				availabilityValue = 1.0
			}
		}
	}

	// Update the metric
	c.metrics.DeviceAvailability.With(prometheus.Labels{
		"device": deviceName,
	}).Set(availabilityValue)
}

// processDeviceMessage processes direct device messages
func (c *Z2MCollector) processDeviceMessage(msg Z2MMessage) {
	deviceName := strings.TrimPrefix(msg.Topic, "zigbee2mqtt/")

	// Handle payload as map
	if payloadMap, ok := msg.Payload.(map[string]interface{}); ok {
		c.updateDeviceMetrics(deviceName, payloadMap)
	}
}

// updateDeviceMetrics updates Prometheus metrics for a device
func (c *Z2MCollector) updateDeviceMetrics(deviceName string, data map[string]interface{}) {
	// Increment seen count
	c.metrics.DeviceSeenCount.With(prometheus.Labels{
		"device": deviceName,
	}).Inc()

	// Update last seen timestamp
	if lastSeen, ok := data["last_seen"].(string); ok {
		if timestamp, err := parseISOTimestamp(lastSeen); err == nil {
			c.metrics.DeviceLastSeen.With(prometheus.Labels{
				"device": deviceName,
			}).Set(float64(timestamp))
		}
	}

	// Update link quality
	if linkQuality, ok := data["linkquality"].(float64); ok {
		c.metrics.DeviceLinkQuality.With(prometheus.Labels{
			"device": deviceName,
		}).Set(linkQuality)
	}

	// Update device state
	if state, ok := data["state"].(string); ok {
		stateValue := 0.0
		if state == "ON" {
			stateValue = 1.0
		}

		c.metrics.DeviceState.With(prometheus.Labels{
			"device": deviceName,
		}).Set(stateValue)
	}

	// Update battery level - check multiple possible field names
	if battery, ok := data["battery"].(float64); ok {
		c.metrics.DeviceBattery.With(prometheus.Labels{
			"device": deviceName,
		}).Set(battery)
	} else if battery, ok := data["battery_level"].(float64); ok {
		c.metrics.DeviceBattery.With(prometheus.Labels{
			"device": deviceName,
		}).Set(battery)
	} else if battery, ok := data["battery_percentage"].(float64); ok {
		c.metrics.DeviceBattery.With(prometheus.Labels{
			"device": deviceName,
		}).Set(battery)
	} else if battery, ok := data["battery_voltage"].(float64); ok {
		// Convert voltage to percentage (typical range: 2.5V-3.3V for Li-ion)
		// This is a reasonable approximation - actual conversion depends on battery chemistry
		batteryPercent := ((battery - 2.5) / 0.8) * 100
		if batteryPercent > 100 {
			batteryPercent = 100
		} else if batteryPercent < 0 {
			batteryPercent = 0
		}

		c.metrics.DeviceBattery.With(prometheus.Labels{
			"device": deviceName,
		}).Set(batteryPercent)
	}
}

// parseISOTimestamp parses an ISO timestamp to Unix timestamp
func parseISOTimestamp(isoTime string) (int64, error) {
	// Remove 'Z' and parse as RFC3339
	isoTime = strings.Replace(isoTime, "Z", "+00:00", 1)

	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		return 0, err
	}

	return t.Unix(), nil
}

// Stop stops the collector
func (c *Z2MCollector) Stop() {
	close(c.done)

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			slog.Error("Failed to close WebSocket connection", "error", err)
		}
	}
}

// min returns the minimum of two time.Duration values
func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}

	return b
}
