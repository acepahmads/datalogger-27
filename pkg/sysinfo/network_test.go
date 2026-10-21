package sysinfo

import (
	"testing"
	"time"
)

func TestSanitizeEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://user:secret123@api.aqms.id/v1/telemetry",
			expected: "https://user:***@api.aqms.id/v1/telemetry",
		},
		{
			input:    "https://api.aqms.id/ingest?token=secrettoken99&station_id=1",
			expected: "https://api.aqms.id/ingest?station_id=1&token=***",
		},
		{
			input:    "http://example.com/api?apiKey=abc-123-xyz&format=json",
			expected: "http://example.com/api?apiKey=***&format=json",
		},
		{
			input:    "mqtt://broker.hivemq.com:1883",
			expected: "mqtt://broker.hivemq.com:1883",
		},
		{
			input:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		actual := SanitizeEndpoint(tc.input)
		if actual != tc.expected {
			t.Errorf("SanitizeEndpoint(%q) = %q, expected %q", tc.input, actual, tc.expected)
		}
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		input    string
		contains string
		notMatch string
	}{
		{
			input:    "HTTP 401 Unauthorized: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xyz",
			notMatch: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xyz",
		},
		{
			input:    "Connection failed with token: my-secret-key-123",
			notMatch: "my-secret-key-123",
		},
	}

	for _, tc := range tests {
		actual := SanitizeError(tc.input)
		if tc.notMatch != "" && actual == tc.input {
			t.Errorf("SanitizeError failed to redact sensitive info: %s", actual)
		}
	}
}

func TestNetworkDiagnosticsCaching(t *testing.T) {
	InvalidateNetworkCache()
	SetNetworkCacheTTL(500 * time.Millisecond)

	diag1 := GetNetworkDiagnostics()
	if diag1.Timestamp.IsZero() {
		t.Errorf("Expected non-zero timestamp")
	}

	// Immediate second call must return cached instance with identical timestamp
	diag2 := GetNetworkDiagnostics()
	if !diag1.Timestamp.Equal(diag2.Timestamp) {
		t.Errorf("Expected cached diagnostics to return identical timestamp, got %v and %v", diag1.Timestamp, diag2.Timestamp)
	}

	// Wait for cache expiry
	time.Sleep(550 * time.Millisecond)
	diag3 := GetNetworkDiagnostics()
	if diag3.Timestamp.Equal(diag1.Timestamp) {
		t.Errorf("Expected fresh probe after TTL expiry")
	}
}

func TestDetectLocalNetwork(t *testing.T) {
	status := DetectLocalNetwork()
	if len(status.Interfaces) == 0 {
		t.Logf("No network interfaces detected (might be isolated test environment)")
	} else {
		t.Logf("Detected %d network interfaces, primary IP: %s, gateway: %s", len(status.Interfaces), status.PrimaryIP, status.Gateway)
	}
}

func TestDetectInternet(t *testing.T) {
	status := DetectInternet()
	t.Logf("Internet status: %s, connected: %v, latency: %.2f ms", status.Status, status.Connected, status.LatencyMs)
	if status.Status != "CONNECTED" && status.Status != "DEGRADED" && status.Status != "DISCONNECTED" {
		t.Errorf("Unexpected internet status: %s", status.Status)
	}
}

func TestMapOutboundStatus(t *testing.T) {
	tests := []struct {
		name              string
		enabled           bool
		rawStatus         string
		successDeliveries int64
		failedDeliveries  int64
		expected          string
	}{
		{"Disabled link", false, "ACTIVE", 100, 0, "DISABLED"},
		{"Healthy link active with successes", true, "ACTIVE", 50, 0, "HEALTHY"},
		{"Healthy link idle but only successes", true, "IDLE", 10, 0, "HEALTHY"},
		{"Failed link with errors only", true, "ERROR", 0, 2746, "FAILED"},
		{"Failed link idle but only failures", true, "IDLE", 0, 15, "FAILED"},
		{"Degraded link with both success and failure", true, "ERROR", 200, 10, "DEGRADED"},
		{"Degraded link idle with both", true, "IDLE", 100, 5, "DEGRADED"},
		{"Unknown state when new destination has no deliveries", true, "IDLE", 0, 0, "UNKNOWN"},
	}

	for _, tc := range tests {
		actual := MapOutboundStatus(tc.enabled, tc.rawStatus, tc.successDeliveries, tc.failedDeliveries)
		if actual != tc.expected {
			t.Errorf("[%s] MapOutboundStatus() = %s, expected %s", tc.name, actual, tc.expected)
		}
	}
}
