package metrics

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test that metrics are properly initialized
	if registry.DeviceLastSeen == nil {
		t.Error("Expected DeviceLastSeen metric to be initialized")
	}

	if registry.DeviceSeenCount == nil {
		t.Error("Expected DeviceSeenCount metric to be initialized")
	}

	if registry.DeviceLinkQuality == nil {
		t.Error("Expected DeviceLinkQuality metric to be initialized")
	}

	if registry.DeviceState == nil {
		t.Error("Expected DeviceState metric to be initialized")
	}

	if registry.DeviceInfo == nil {
		t.Error("Expected DeviceInfo metric to be initialized")
	}

	if registry.DeviceAvailability == nil {
		t.Error("Expected DeviceAvailability metric to be initialized")
	}

	if registry.BridgeState == nil {
		t.Error("Expected BridgeState metric to be initialized")
	}

	if registry.BridgeEventsTotal == nil {
		t.Error("Expected BridgeEventsTotal metric to be initialized")
	}

	if registry.WebSocketConnectionStatus == nil {
		t.Error("Expected WebSocketConnectionStatus metric to be initialized")
	}

	if registry.WebSocketMessagesTotal == nil {
		t.Error("Expected WebSocketMessagesTotal metric to be initialized")
	}

	if registry.WebSocketReconnectsTotal == nil {
		t.Error("Expected WebSocketReconnectsTotal metric to be initialized")
	}
}
