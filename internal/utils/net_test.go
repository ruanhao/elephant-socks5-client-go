package utils

import "testing"

func TestGetOutboundIP(t *testing.T) {
	ip := GetOutboundIP()
	if ip == "" {
		t.Error("Expected a non-empty IP address, got an empty string")
	}
	t.Logf("Outbound IP address: %s", ip)
}
