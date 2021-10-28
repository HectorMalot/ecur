package ecur

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests only work with an ECU-R device connected via WiFi on the specified IP
func TestClient(t *testing.T) {
	ip, ok := os.LookupEnv("APS_IP")
	if !ok {
		t.Skip("Skipping TestClient: no APS_IP variable provided")
	}
	c, err := NewClient(ip, DefaultPort, "Europe/Amsterdam")
	require.NoError(t, err)

	// Valid port should connect
	err = c.Connect()
	require.NoError(t, err)
	defer c.Close()

	// Should be able to download ECU data
	info, err := c.GetECUInfo()
	require.NoError(t, err)
	require.Contains(t, info.Version, "ECU_R_")

	// Should be able to download inverter data
	invInfo, err := c.GetInverterInfo()
	require.NoError(t, err)
	require.Equal(t, true, invInfo.Inverters[0].Online)
}

func TestClientGetData(t *testing.T) {
	ip, ok := os.LookupEnv("APS_IP")
	if !ok {
		t.Skip("Skipping TestClientGetData: no APS_IP variable provided")
	}
	c, err := NewClient(ip, DefaultPort, "Europe/Amsterdam")
	require.NoError(t, err)

	// E2E call should succeed
	_, err = c.GetData()
	require.NoError(t, err)
}
