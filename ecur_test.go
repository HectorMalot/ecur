package ecur

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestECUInfoParser(t *testing.T) {
	// at night
	raw := []byte{65, 80, 83, 49, 49, 48, 48, 57, 52, 48, 48, 48, 49, 50, 49, 54, 48, 48, 48, 48, 49, 49, 49, 49, 49, 48, 49, 0, 0, 166, 159, 0, 0, 0, 0, 0, 0, 1, 140, 208, 208, 208, 208, 208, 208, 208, 0, 2, 0, 0, 49, 48, 48, 49, 50, 69, 67, 85, 95, 82, 95, 49, 46, 50, 46, 49, 56, 48, 48, 57, 69, 116, 99, 47, 71, 77, 84, 45, 56, 128, 151, 27, 1, 164, 227, 0, 0, 0, 0, 0, 0, 69, 78, 68, 10}
	info, err := NewECUInfo(raw)
	require.NoError(t, err)
	require.Equal(t, "ECU_R_1.2.18", info.Version)
	require.Equal(t, "216000011111", info.EcuID)
	require.Equal(t, 0, info.InvertersOnline)
	require.Equal(t, 2, info.InvertersRegistered)
	require.Equal(t, 4265500, info.LifetimeEnergy)
	require.Equal(t, 3960, info.TodayEnergy)
	require.Equal(t, 0, info.LastPower)

	// during day
	raw = []byte{65, 80, 83, 49, 49, 48, 48, 57, 52, 48, 48, 48, 49, 50, 49, 54, 48, 48, 48, 48, 49, 49, 49, 49, 49, 48, 49, 0, 0, 166, 243, 0, 0, 1, 36, 0, 0, 0, 69, 208, 208, 208, 208, 208, 208, 208, 0, 2, 0, 2, 49, 48, 48, 49, 50, 69, 67, 85, 95, 82, 95, 49, 46, 50, 46, 49, 57, 48, 48, 57, 69, 116, 99, 47, 71, 77, 84, 45, 56, 128, 151, 27, 1, 164, 227, 0, 0, 0, 0, 0, 0, 69, 78, 68, 10}
	info, err = NewECUInfo(raw)
	require.NoError(t, err)
	require.Equal(t, "ECU_R_1.2.19", info.Version)
	require.Equal(t, "216000011111", info.EcuID)
	require.Equal(t, 2, info.InvertersOnline)
	require.Equal(t, 2, info.InvertersRegistered)
	require.Equal(t, 4273900, info.LifetimeEnergy)
	require.Equal(t, 690, info.TodayEnergy)
	require.Equal(t, 292, info.LastPower)
}

func TestArrayInfoParser(t *testing.T) {
	// at night
	raw := []byte{65, 80, 83, 49, 49, 48, 48, 55, 53, 48, 48, 48, 50, 48, 48, 48, 49, 0, 2, 32, 33, 16, 24, 34, 82, 16, 128, 16, 0, 3, 0, 0, 0, 48, 51, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 2, 0, 3, 128, 16, 0, 3, 0, 1, 0, 48, 51, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 2, 0, 2, 69, 78, 68, 10}
	info, err := NewArrayInfo(raw, "Europe/Amsterdam")
	require.NoError(t, err)
	require.Equal(t, false, info.Inverters[0].Online)

	// during day
	raw = []byte{65, 80, 83, 49, 49, 48, 48, 55, 53, 48, 48, 48, 50, 48, 48, 48, 49, 0, 2, 32, 33, 16, 32, 20, 24, 5, 128, 16, 0, 3, 0, 0, 1, 48, 51, 1, 243, 0, 119, 0, 57, 0, 228, 0, 56, 0, 60, 0, 60, 128, 16, 0, 3, 0, 1, 1, 48, 51, 1, 243, 0, 118, 0, 55, 0, 229, 0, 55, 0, 57, 0, 56, 69, 78, 68, 10}
	info, err = NewArrayInfo(raw, "Europe/Amsterdam")
	require.NoError(t, err)
	require.Equal(t, true, info.Inverters[0].Online)
}

func TestInverterSignalParser(t *testing.T) {
	// during day
	raw := []byte{65, 80, 83, 49, 49, 48, 48, 51, 50, 48, 48, 51, 48, 48, 48, 128, 16, 0, 3, 0, 0, 213, 128, 16, 0, 3, 0, 1, 223, 69, 78, 68, 10}
	info, err := NewInverterSignalinfo(raw)
	require.NoError(t, err)
	require.Equal(t, 213, info.Inverters[0].Signal)
	require.Equal(t, 223, info.Inverters[1].Signal)
}

func TestByteSliceToString(t *testing.T) {
	require.Equal(t, "0123EF", byteSliceToString([]byte{0x01, 0x23, 0xef}))
}

func TestTimestampParser(t *testing.T) {
	body := []byte{0x20, 0x21, 0x10, 0x28, 0x10, 0x00, 0x01}
	ts, err := binToTimestamp(body, "Europe/Amsterdam")
	require.NoError(t, err)
	require.Equal(t, 2021, ts.Year())
	require.Equal(t, time.October, ts.Month())
	require.Equal(t, 28, ts.Day())
	require.Equal(t, 10, ts.Hour())
	require.Equal(t, 0, ts.Minute())
	require.Equal(t, 01, ts.Second())
}
