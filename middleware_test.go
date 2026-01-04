package wakeonlan

import (
	"context"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalCaddyfile(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedMAC   string
		expectedBcast string
		expectedTime  time.Duration
		shouldErr     bool
	}{
		{
			name:          "One argument",
			input:         "wake_on_lan CC:C4:45:32:7A:51",
			expectedMAC:   "CC:C4:45:32:7A:51",
			expectedBcast: "255.255.255.255:9",
			expectedTime:  10 * time.Minute,
			shouldErr:     false,
		},
		{
			name:          "Two arguments",
			input:         "wake_on_lan CC:C4:45:32:7A:51 192.168.1.255:9",
			expectedMAC:   "CC:C4:45:32:7A:51",
			expectedBcast: "192.168.1.255:9",
			expectedTime:  10 * time.Minute,
			shouldErr:     false,
		},
		{
			name:          "Three arguments",
			input:         "wake_on_lan CC:C4:45:32:7A:51 192.168.1.255:9 5m",
			expectedMAC:   "CC:C4:45:32:7A:51",
			expectedBcast: "192.168.1.255:9",
			expectedTime:  5 * time.Minute,
			shouldErr:     false,
		},
		{
			name:          "Invalid Duration",
			input:         "wake_on_lan CC:C4:45:32:7A:51 192.168.1.255:9 invalid",
			expectedMAC:   "",
			expectedBcast: "",
			expectedTime:  0,
			shouldErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := caddyfile.NewTestDispenser(tt.input)
			m := Middleware{}
			err := m.UnmarshalCaddyfile(d)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if m.MAC != tt.expectedMAC {
				t.Errorf("Expected MAC %s, got %s", tt.expectedMAC, m.MAC)
			}
			if m.BroadcastAddress != tt.expectedBcast {
				t.Errorf("Expected BroadcastAddress %s, got %s", tt.expectedBcast, m.BroadcastAddress)
			}
			if time.Duration(m.Timeout) != tt.expectedTime {
				t.Errorf("Expected Timeout %v, got %v", tt.expectedTime, time.Duration(m.Timeout))
			}
		})
	}
}

func TestProvisionDefaultTimeout(t *testing.T) {
	// Setup a middleware with zero timeout (simulating default JSON unmarshal)
	m := Middleware{
		MAC:              "00:00:00:00:00:00",
		BroadcastAddress: "127.0.0.1:9",
	}

	// Create a dummy context
	ctx, cancel := caddy.NewContext(caddy.Context{Context: context.Background()})
	defer cancel()

	// Run Provision
	// Note: We expect this to succeed because 127.0.0.1:9 is a valid address to "dial" for UDP even if nothing is listening,
	// unless there's a permission error or similar.
	err := m.Provision(ctx)
	if err != nil {
		t.Fatalf("Provision failed: %v", err)
	}

	// Check if Timeout was set to default 10m
	expected := caddy.Duration(10 * time.Minute)
	if m.Timeout != expected {
		t.Errorf("Expected default timeout %v, got %v", expected, m.Timeout)
	}
}
