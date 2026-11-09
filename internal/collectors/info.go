package collectors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/d0ugal/promexporter/tracing"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

// bridgeInfo is the subset of the bridge/info payload the exporter uses.
type bridgeInfo struct {
	Version         string `json:"version"`
	Commit          string `json:"commit"`
	PermitJoin      bool   `json:"permit_join"`
	RestartRequired bool   `json:"restart_required"`
	Coordinator     *struct {
		IEEEAddress string `json:"ieee_address"`
		Type        string `json:"type"`
		Meta        struct {
			// Z-Stack reports a numeric revision, ember a string.
			Revision interface{} `json:"revision"`
		} `json:"meta"`
	} `json:"coordinator"`
	Network *struct {
		Channel float64 `json:"channel"`
		PanID   float64 `json:"pan_id"`
	} `json:"network"`
	ZigbeeHerdsman struct {
		Version string `json:"version"`
	} `json:"zigbee_herdsman"`
	ZigbeeHerdsmanConverters struct {
		Version string `json:"version"`
	} `json:"zigbee_herdsman_converters"`
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}

	return 0
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}

	return s
}

// formatRevision renders a coordinator revision without float noise, so
// 20221226 stays "20221226" rather than "2.0221226e+07".
func formatRevision(v interface{}) string {
	switch r := v.(type) {
	case nil:
		return "unknown"
	case float64:
		return fmt.Sprintf("%.0f", r)
	case string:
		return orUnknown(r)
	default:
		return fmt.Sprintf("%v", r)
	}
}

// processBridgeInfoMessage processes bridge/info messages, which describe the
// Zigbee2MQTT version, coordinator adapter and network settings.
func (c *Z2MCollector) processBridgeInfoMessage(ctx context.Context, msg Z2MMessage) {
	tracer := c.app.GetTracer()

	var span *tracing.CollectorSpan

	if tracer != nil && tracer.IsEnabled() {
		span = tracer.NewCollectorSpan(ctx, "z2m-collector", "process-bridge-info")
		span.SetAttributes(attribute.String("message.type", "bridge/info"))

		defer span.End()
	}

	raw, err := json.Marshal(msg.Payload)
	if err != nil {
		if span != nil {
			span.RecordError(err, attribute.String("operation", "marshal_payload"))
		}

		return
	}

	var info bridgeInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		slog.Debug("Could not parse bridge/info message", "error", err)

		if span != nil {
			span.RecordError(err, attribute.String("operation", "parse_info"))
		}

		return
	}

	m := c.metrics.BridgeInfo
	noLabels := prometheus.Labels{}

	// Reset the info metrics so an upgrade replaces the old series instead
	// of leaving both reporting 1.
	m.Info.Reset()
	m.Info.With(prometheus.Labels{
		"version":                            orUnknown(info.Version),
		"commit":                             orUnknown(strings.TrimSpace(info.Commit)),
		"zigbee_herdsman_version":            orUnknown(info.ZigbeeHerdsman.Version),
		"zigbee_herdsman_converters_version": orUnknown(info.ZigbeeHerdsmanConverters.Version),
	}).Set(1)

	if info.Coordinator != nil {
		m.CoordinatorInfo.Reset()
		m.CoordinatorInfo.With(prometheus.Labels{
			"type":         orUnknown(info.Coordinator.Type),
			"ieee_address": orUnknown(info.Coordinator.IEEEAddress),
			"revision":     formatRevision(info.Coordinator.Meta.Revision),
		}).Set(1)
	}

	if info.Network != nil {
		m.NetworkChannel.With(noLabels).Set(info.Network.Channel)
		m.NetworkPanID.With(noLabels).Set(info.Network.PanID)
	}

	m.PermitJoin.With(noLabels).Set(boolToFloat(info.PermitJoin))
	m.RestartRequired.With(noLabels).Set(boolToFloat(info.RestartRequired))

	if span != nil {
		span.SetAttributes(
			attribute.String("bridge.version", info.Version),
			attribute.Bool("bridge.permit_join", info.PermitJoin),
		)
	}
}
