package metrics

import (
	promexporter_metrics "github.com/d0ugal/promexporter/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// BridgeInfoMetrics holds the metrics populated from Zigbee2MQTT's bridge/info topic.
type BridgeInfoMetrics struct {
	Info            *prometheus.GaugeVec
	CoordinatorInfo *prometheus.GaugeVec
	NetworkChannel  *prometheus.GaugeVec
	NetworkPanID    *prometheus.GaugeVec
	PermitJoin      *prometheus.GaugeVec
	RestartRequired *prometheus.GaugeVec
}

var (
	bridgeInfoLabels      = []string{"version", "commit", "zigbee_herdsman_version", "zigbee_herdsman_converters_version"}
	coordinatorInfoLabels = []string{"type", "ieee_address", "revision"}
)

func newBridgeInfoMetrics(baseRegistry *promexporter_metrics.Registry, factory promauto.Factory) *BridgeInfoMetrics {
	gauge := func(name, help string, labels []string) *prometheus.GaugeVec {
		baseRegistry.AddMetricInfo(name, help, labels)

		return factory.NewGaugeVec(prometheus.GaugeOpts{Name: name, Help: help}, labels)
	}

	none := []string{}

	return &BridgeInfoMetrics{
		Info: gauge("zigbee2mqtt_bridge_info",
			"Zigbee2MQTT version information (always 1, from bridge/info)", bridgeInfoLabels),
		CoordinatorInfo: gauge("zigbee2mqtt_coordinator_info",
			"Zigbee coordinator adapter type and firmware revision (always 1, from bridge/info)", coordinatorInfoLabels),
		NetworkChannel: gauge("zigbee2mqtt_network_channel",
			"Zigbee channel the network is running on (from bridge/info)", none),
		NetworkPanID: gauge("zigbee2mqtt_network_pan_id",
			"Zigbee network PAN ID (from bridge/info)", none),
		PermitJoin: gauge("zigbee2mqtt_bridge_permit_join",
			"Whether new devices are allowed to join the network (1=yes, 0=no, from bridge/info)", none),
		RestartRequired: gauge("zigbee2mqtt_bridge_restart_required",
			"Whether Zigbee2MQTT needs a restart to apply configuration changes (1=yes, 0=no, from bridge/info)", none),
	}
}
