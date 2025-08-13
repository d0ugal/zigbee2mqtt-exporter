# Zigbee2MQTT Exporter

A Prometheus exporter for Zigbee2MQTT that connects to the WebSocket API and exports device metrics with device type classification.

## Features

- **Real-time metrics**: Connects to Zigbee2MQTT WebSocket API for real-time device data
- **Device type classification**: Automatically categorizes devices as Router, EndDevice, or Coordinator
- **Device monitoring**: Tracks device last seen timestamps, seen counts, link quality, and state
- **Bridge monitoring**: Monitors bridge state, events, and permit join status
- **Connection monitoring**: Monitors WebSocket connection status and reconnection attempts
- **Web UI**: Built-in web interface for monitoring exporter status
- **Prometheus integration**: Exports metrics in Prometheus format

## Device Types

The exporter automatically categorizes devices based on their Zigbee network role:

- **Router**: Powered devices that extend the mesh network (lights, plugs, switches, etc.)
- **EndDevice**: Battery-powered devices that don't route traffic (sensors, switches, etc.)
- **Coordinator**: The main Zigbee2MQTT bridge device
- **GreenPower**: Special Zigbee Green Power devices (energy harvesting devices)

This classification enables different alerting strategies:
- **Routers going offline**: Critical - affects network connectivity
- **EndDevices going offline**: Less critical - may be due to battery depletion or sleep mode
- **GreenPower devices going offline**: Less critical - may be due to energy harvesting limitations
- **Coordinator offline**: Critical - entire network is down

## Metrics

### Device Metrics
- `zigbee2mqtt_device_last_seen_timestamp{device}` - Timestamp when device was last seen
- `zigbee2mqtt_device_seen_total{device}` - Number of times device has been seen
- `zigbee2mqtt_device_link_quality{device}` - Device link quality (0-255)
- `zigbee2mqtt_device_power_state{device}` - Device power state (1=ON, 0=OFF)
- `zigbee2mqtt_device_up{device}` - Device availability status (1=online, 0=offline)

### Device Info Metric
- `zigbee2mqtt_device_info{device,type,power_source,manufacturer,model_id,supported,disabled,interview_state}` - Device information (always 1, used for joining)

### Bridge Metrics
- `zigbee2mqtt_bridge_state` - Bridge state (1=online, 0=offline)
- `zigbee2mqtt_bridge_events_total{event_type}` - Total number of bridge events

### Connection Metrics
- `zigbee2mqtt_websocket_connection_status` - WebSocket connection status (1=connected, 0=disconnected)
- `zigbee2mqtt_websocket_messages_total{topic}` - Total number of WebSocket messages received
- `zigbee2mqtt_websocket_reconnects_total` - Total number of WebSocket reconnections

## Querying Examples

### Get all battery-powered devices that are offline
```promql
zigbee2mqtt_device_last_seen_timestamp * on(device) group_left(type, power_source) zigbee2mqtt_device_info
  * on(device) group_left() (zigbee2mqtt_device_info{power_source="Battery"})
```

### Get link quality for all routers
```promql
zigbee2mqtt_device_link_quality * on(device) group_left(type, power_source) zigbee2mqtt_device_info
  * on(device) group_left() (zigbee2mqtt_device_info{type="Router"})
```

### Count devices by type
```promql
count by (type) (zigbee2mqtt_device_info)
```

### Count devices by power source
```promql
count by (power_source) (zigbee2mqtt_device_info)
```

### Get unsupported devices
```promql
zigbee2mqtt_device_info{supported="false"}
```

## Configuration

### Environment Variables

- `Z2M_EXPORTER_SERVER_HOST` - HTTP server host (default: 0.0.0.0)
- `Z2M_EXPORTER_SERVER_PORT` - HTTP server port (default: 8087)
- `Z2M_EXPORTER_LOG_LEVEL` - Log level (default: info)
- `Z2M_EXPORTER_LOG_FORMAT` - Log format (default: json)
- `Z2M_EXPORTER_WEBSOCKET_URL` - Zigbee2MQTT WebSocket URL (default: ws://localhost:8081/api)

## Usage

### Docker Compose

The exporter is configured in the main docker-compose.yaml and will build automatically.

### Manual Build

```bash
# Build the image
docker build -t zigbee2mqtt-exporter .

# Run with environment variables
docker run -p 8087:8087 \
  -e Z2M_EXPORTER_WEBSOCKET_URL=ws://localhost:8081/api \
  zigbee2mqtt-exporter
```

## Endpoints

- `/metrics` - Prometheus metrics endpoint
- `/health` - Health check endpoint
- `/` - Web UI interface

## Alerting Examples

### Critical Router Offline Alert
```yaml
- alert: ZigbeeRouterOffline
  expr: zigbee2mqtt_device_last_seen_timestamp{type="Router"} < (time() - 300)
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "Zigbee router {{ $labels.device }} is offline"
    description: "Router {{ $labels.device }} has been offline for more than 5 minutes"
```

### EndDevice Offline Warning
```yaml
- alert: ZigbeeEndDeviceOffline
  expr: zigbee2mqtt_device_last_seen_timestamp{type="EndDevice"} < (time() - 3600)
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "Zigbee end device {{ $labels.device }} is offline"
    description: "End device {{ $labels.device }} has been offline for more than 1 hour (may be battery powered)"
```

### Bridge Offline Critical Alert
```yaml
- alert: ZigbeeBridgeOffline
  expr: zigbee2mqtt_bridge_state == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "Zigbee2MQTT bridge is offline"
    description: "The Zigbee2MQTT bridge has gone offline"
```

## Testing

Use the included message capture tool to test WebSocket connectivity:

```bash
cd tools
go run capture_messages.go ws://localhost:8081/api
```

## Development

### Building Locally

```bash
go mod download
go build -o zigbee2mqtt-exporter ./cmd
```

### Running Tests

```bash
go test ./...
```

## Architecture

The exporter follows a modular design:

- **cmd/main.go** - Application entry point
- **internal/config** - Configuration management
- **internal/logging** - Logging setup
- **internal/metrics** - Prometheus metrics definitions
- **internal/collectors** - WebSocket data collection with device type caching
- **internal/server** - HTTP server and web UI

## License

This project follows the same license as the parent repository.
