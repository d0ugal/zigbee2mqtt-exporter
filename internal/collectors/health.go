package collectors

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/d0ugal/promexporter/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

// bridgeHealth is the payload Zigbee2MQTT publishes on bridge/health.
// Sections are pointers so that a missing section leaves its metrics untouched.
type bridgeHealth struct {
	ResponseTime float64 `json:"response_time"`
	Process      *struct {
		UptimeSec    float64 `json:"uptime_sec"`
		MemoryUsedMB float64 `json:"memory_used_mb"`
	} `json:"process"`
	MQTT *struct {
		Connected bool    `json:"connected"`
		Queued    float64 `json:"queued"`
		Published float64 `json:"published"`
		Received  float64 `json:"received"`
	} `json:"mqtt"`
	Devices map[string]struct {
		Messages              float64 `json:"messages"`
		MessagesPerSec        float64 `json:"messages_per_sec"`
		LeaveCount            float64 `json:"leave_count"`
		NetworkAddressChanges float64 `json:"network_address_changes"`
	} `json:"devices"`
}

// rememberIEEE records the IEEE address to friendly name mapping for a device.
func (c *Z2MCollector) rememberIEEE(device BridgeDevice) {
	if device.IEEEAddress != "" && device.FriendlyName != "" {
		c.ieeeToName[device.IEEEAddress] = device.FriendlyName
	}
}

// processBridgeHealthMessage processes bridge/health messages, which carry
// per-device message counters plus Zigbee2MQTT process and MQTT statistics.
func (c *Z2MCollector) processBridgeHealthMessage(ctx context.Context, msg Z2MMessage) {
	tracer := c.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "z2m-collector", "process-bridge-health")
		span.SetAttributes(attribute.String("message.type", "bridge/health"))

		defer span.End()
	}

	raw, err := json.Marshal(msg.Payload)
	if err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("operation", "marshal_payload"))
		}

		return
	}

	var health bridgeHealth
	if err := json.Unmarshal(raw, &health); err != nil {
		slog.Debug("Could not parse bridge/health message", "error", err)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_health"))
		}

		return
	}

	m := c.metrics.Health
	noLabels := prometheus.Labels{}

	if health.ResponseTime > 0 {
		// response_time is a Unix timestamp in milliseconds
		m.LastReport.With(noLabels).Set(health.ResponseTime / 1000)
	}

	if health.Process != nil {
		m.ProcessUptime.With(noLabels).Set(health.Process.UptimeSec)
		m.ProcessMemoryUsed.With(noLabels).Set(health.Process.MemoryUsedMB * 1024 * 1024)
	}

	if health.MQTT != nil {
		connected := 0.0
		if health.MQTT.Connected {
			connected = 1.0
		}

		m.MQTTConnected.With(noLabels).Set(connected)
		m.MQTTQueued.With(noLabels).Set(health.MQTT.Queued)
		m.MQTTPublished.With(noLabels).Set(health.MQTT.Published)
		m.MQTTReceived.With(noLabels).Set(health.MQTT.Received)
	}

	var updated, unknown int

	for ieee, stats := range health.Devices {
		name, ok := c.ieeeToName[ieee]
		if !ok {
			// bridge/devices has not arrived yet (or the device was just
			// removed); skip rather than emit a series labelled by IEEE
			// address that would be orphaned once the name is known.
			unknown++

			continue
		}

		labels := prometheus.Labels{"device": name}
		m.DeviceMessages.With(labels).Set(stats.Messages)
		m.DeviceMessagesPerSecond.With(labels).Set(stats.MessagesPerSec)
		m.DeviceLeaveCount.With(labels).Set(stats.LeaveCount)
		m.DeviceNetworkAddressChanges.With(labels).Set(stats.NetworkAddressChanges)

		updated++
	}

	if unknown > 0 {
		slog.Debug("Skipped bridge/health devices with unknown IEEE address", "count", unknown)
	}

	if span != nil {
		span.SetAttributes(
			attribute.Int("health.devices_updated", updated),
			attribute.Int("health.devices_unknown", unknown),
		)
	}
}
