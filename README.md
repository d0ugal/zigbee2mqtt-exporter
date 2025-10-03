# Zigbee2MQTT Exporter

A Prometheus exporter for Zigbee2MQTT that connects to the WebSocket API and exports device metrics with device type classification.

**Image**: `ghcr.io/d0ugal/zigbee2mqtt-exporter:v2.14.0`

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
- `zigbee2mqtt_device_ota_update_available{device}` - Device OTA update availability (1=available, 0=not_available)
- `zigbee2mqtt_device_current_firmware_version{device,firmware_version}` - Device current firmware version (always 1, used for joining)
- `zigbee2mqtt_device_available_firmware_version{device,firmware_version}` - Device available firmware version (always 1, used for joining)

### Bridge Metrics
- `zigbee2mqtt_bridge_state` - Bridge state (1=online, 0=offline)
- `zigbee2mqtt_bridge_events_total{event_type}` - Total number of bridge events

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
    image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v2.14.0
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
    image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v2.14.0
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
        image: ghcr.io/d0ugal/zigbee2mqtt-exporter:v2.14.0
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
zigbee2mqtt_device_ota_update_available == 1
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