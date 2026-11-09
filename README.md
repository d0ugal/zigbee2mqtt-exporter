# Zigbee2MQTT Exporter

A Prometheus exporter for Zigbee2MQTT that connects to the WebSocket API and exports device metrics with device type classification.

**Image**: `ghcr.io/d0ugal/zigbee2mqtt-exporter:v3.5.55`

## Metrics

### Device Metrics
- `zigbee2mqtt_device_last_seen_timestamp{device}` - Timestamp when device was last seen
- `zigbee2mqtt_device_seen_total{device}` - Number of times device has been seen
- `zigbee2mqtt_device_link_quality{device}` - Device link quality (0-255)
- `zigbee2mqtt_device_power_state{device}` - Device power state (1=ON, 0=OFF)
- `zigbee2mqtt_device_up{device}` - Device availability status (1=online, 0=offline)

### Device Info Metric
- `zigbee2mqtt_device_info{device,type,power_source,manufacturer,model_id,supported,disabled,interview_state,software_build_id,date_code}` - Device information (always 1, used for joining)

### OTA Update Metrics
- `zigbee2mqtt_device_ota{device,state}` - Device OTA update state (state=idle|available|scheduled|updating, value is 1 when device is in that state)
- `zigbee2mqtt_device_current_firmware_version{device,firmware_version}` - Installed OTA file version, from the device's `update.installed_version` (always 1, used for joining)
- `zigbee2mqtt_device_ota_progress_percent{device}` - Progress of an in-progress OTA update (0-100), only present while updating
- `zigbee2mqtt_device_ota_remaining_seconds{device}` - Estimated seconds left of an in-progress OTA update, only present while updating
- `zigbee2mqtt_device_available_firmware_version{device,firmware_version}` - Latest OTA file version, from the device's `update.latest_version` (always 1, used for joining)

### Bridge Metrics
- `zigbee2mqtt_bridge_state` - Bridge state (1=online, 0=offline)
- `zigbee2mqtt_bridge_events_total{event_type}` - Total number of bridge events

### Health Metrics (from `bridge/health`)
Zigbee2MQTT publishes `bridge/health` periodically (every 10 minutes by default). Its counters reset when Zigbee2MQTT restarts, so they are exported as gauges; `rate()` and `increase()` handle the resets.

- `zigbee2mqtt_device_health_messages{device}` - Messages received from the device since Zigbee2MQTT started
- `zigbee2mqtt_device_health_messages_per_second{device}` - Average messages per second since Zigbee2MQTT started
- `zigbee2mqtt_device_health_leave_count{device}` - Times the device has left the network
- `zigbee2mqtt_device_health_network_address_changes{device}` - Times the device's network address has changed
- `zigbee2mqtt_process_uptime_seconds` - Zigbee2MQTT process uptime
- `zigbee2mqtt_process_memory_used_bytes` - Zigbee2MQTT process memory usage
- `zigbee2mqtt_mqtt_connected` - Whether Zigbee2MQTT is connected to the MQTT broker (1/0)
- `zigbee2mqtt_mqtt_queued_messages` - MQTT messages waiting to be published
- `zigbee2mqtt_mqtt_published_messages` - MQTT messages published since Zigbee2MQTT started
- `zigbee2mqtt_mqtt_received_messages` - MQTT messages received since Zigbee2MQTT started
- `zigbee2mqtt_health_last_report_timestamp_seconds` - When the last `bridge/health` report was received

### Bridge Info Metrics (from `bridge/info`)
- `zigbee2mqtt_bridge_info{version,commit,zigbee_herdsman_version,zigbee_herdsman_converters_version}` - Zigbee2MQTT version information (always 1)
- `zigbee2mqtt_coordinator_info{type,ieee_address,revision}` - Coordinator adapter type and firmware revision (always 1)
- `zigbee2mqtt_network_channel` - Zigbee channel the network is running on
- `zigbee2mqtt_network_pan_id` - Zigbee network PAN ID
- `zigbee2mqtt_bridge_permit_join` - Whether new devices are allowed to join (1/0)
- `zigbee2mqtt_bridge_restart_required` - Whether Zigbee2MQTT needs a restart to apply configuration changes (1/0)

### Connection Metrics
- `zigbee2mqtt_websocket_connection_status` - WebSocket connection status (1=connected, 0=disconnected)
- `zigbee2mqtt_websocket_messages_total{topic}` - Total number of WebSocket messages received
- `zigbee2mqtt_websocket_reconnects_total` - Total number of WebSocket reconnections

### Endpoints
- `GET /`: Web UI interface
- `GET /metrics`: Prometheus metrics endpoint
- `GET /health`: Health check endpoint

## Quick Start

### Docker Compose

```yaml
version: '3.8'
services:
  zigbee2mqtt-exporter:
    image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v3.5.55
    ports:
      - "8087:8087"
    environment:
      - Z2M_EXPORTER_WEBSOCKET_URL=ws://localhost:8081/api
    restart: unless-stopped
```

1. Update the WebSocket URL to point to your Zigbee2MQTT instance
2. Run: `docker-compose up -d`
3. Access metrics: `curl http://localhost:8087/metrics`

## Configuration

### Environment Variables

- `Z2M_EXPORTER_SERVER_HOST` - HTTP server host (default: 0.0.0.0)
- `Z2M_EXPORTER_SERVER_PORT` - HTTP server port (default: 8087)
- `Z2M_EXPORTER_LOG_LEVEL` - Log level (default: info)
- `Z2M_EXPORTER_LOG_FORMAT` - Log format (default: json)
- `Z2M_EXPORTER_WEBSOCKET_URL` - Zigbee2MQTT WebSocket URL (default: ws://localhost:8081/api)

## Deployment

### Docker Compose (Environment Variables)

```yaml
version: '3.8'
services:
  zigbee2mqtt-exporter:
    image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v3.5.55
    ports:
      - "8087:8087"
    environment:
      - Z2M_EXPORTER_WEBSOCKET_URL=ws://localhost:8081/api
      - Z2M_EXPORTER_LOG_LEVEL=info
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zigbee2mqtt-exporter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: zigbee2mqtt-exporter
  template:
    metadata:
      labels:
        app: zigbee2mqtt-exporter
    spec:
      containers:
      - name: zigbee2mqtt-exporter
        image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v3.5.55
        ports:
        - containerPort: 8087
        env:
        - name: Z2M_EXPORTER_WEBSOCKET_URL
          value: "ws://zigbee2mqtt:8081/api"
        - name: Z2M_EXPORTER_LOG_LEVEL
          value: "info"
```

## Prometheus Integration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'zigbee2mqtt-exporter'
    static_configs:
      - targets: ['zigbee2mqtt-exporter:8087']
```

## Device Types

The exporter automatically categorizes devices based on their Zigbee network role:

- **Router**: Powered devices that extend the mesh network (lights, plugs, switches, etc.)
- **EndDevice**: Battery-powered devices that don't route traffic (sensors, switches, etc.)
- **Coordinator**: The main Zigbee2MQTT bridge device
- **GreenPower**: Special Zigbee Green Power devices (energy harvesting devices)

## Example PromQL Queries

### Count devices by type
```promql
count by (type) (zigbee2mqtt_device_info)
```

### Get devices with OTA updates available
```promql
zigbee2mqtt_device_ota{state="available"} == 1
```

### Get devices currently updating
```promql
zigbee2mqtt_device_ota{state="updating"} == 1
```

### Get count of devices in each OTA state
```promql
count by (state) (zigbee2mqtt_device_ota == 1)
```

### Get unsupported devices
```promql
zigbee2mqtt_device_info{supported="false"}
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## License

This project follows the same license as the parent repository.