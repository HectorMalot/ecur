package ecur

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
)

const (
	DefaultPort             = 8899
	DefaultTz               = "UTC"
	CmdECUInfo              = "APS1100160001END\n"
	CmdInverterInfoPrefix   = "APS1100280002"
	CmdInverterInfoSuffix   = "END\n"
	CmdInverterSignalPrefix = "APS1100280030"
	CmdInverterSignalSuffix = "END\n"
	CmdGetEnergyPrefix      = "APS1100390004"
	CmdGetEnergyWeekSuffix  = "END00END\n"
	CmdGetEnergyMonthSuffix = "END01END\n"
	CmdGetEnergyYearSuffix  = "END02END\n"
)

type ECUResponse struct {
	ECUInfo            ECUInfo
	ArrayInfo          ArrayInfo
	InverterSignalInfo InverterSignalInfo
}

type ECUInfo struct {
	EcuID               string
	Version             string
	InvertersRegistered int
	InvertersOnline     int
	EthernetMac         string
	WirelessMac         string

	LifetimeEnergy int // in Wh
	TodayEnergy    int // in Wh
	LastPower      int // in W

	Raw []byte
}

/*
NewECUInfo parses the ECUInfo call into a struct

Decoded explanation
---------------------
 0- 2 : APS          = Mark start of datastream
 3- 4 : 11           = Answer notation
 4- 8 : 0094         = Datalength
 9-12 : 0001         = commandnumber
13-24 : 216000026497 = ECU_R nummer
25-26 : 01           = number of inverters online
27-30 : 42896        = Lifetime energy (kWh)/10
31-34 : 00 00 00 00  = LastSystemPower kW/100
35-38 : 202          = CurrentDayEnergy (/100)
39-45 : 7xd0         = LastTimeConnectEMA
46-47 : 8            = number of inverters registered
48-49 : 0            = number of inverters online
50-51 : 10           = EcuChannel
52-54 : 014          = VersionLEN => VL
55-55+VL          : ECU_R_1.2.13 = Version
56+VL-57+VL       : 009          = TimeZoneLen => TL
58+VL-57+VL+TL    : Etc/GMT-8    = Timezone server always indicated at -8 hours
58+VL+TL-63+VL+TL : 80 97 1b 01 5d 1e = EthernetMAC
64+VL+TL-69+VL+TL : 00 00 00 00 00 00 = WirelessMAC //Shoud be but there is a bugin firmware
70+VL+TL-73+VL+TL : END\n             = SignatureStop Marks end of datastream
*/
func NewECUInfo(raw []byte) (ECUInfo, error) {
	// Validation
	err := validateLength(raw)
	if err != nil {
		return ECUInfo{Raw: raw}, err
	}

	// Version
	verLength, err := strconv.Atoi(string(raw[52:55]))
	if err != nil {
		return ECUInfo{Raw: raw}, fmt.Errorf("could not parse version from body: %w", ErrMalformedBody)
	}
	version := string(raw[55 : 55+verLength])

	// TZ length
	tzLength, err := strconv.Atoi(string(raw[55+verLength : 55+verLength+3]))
	if err != nil {
		return ECUInfo{Raw: raw}, fmt.Errorf("could not parse timezone length from body: %w", ErrMalformedBody)
	}

	// Return struct
	return ECUInfo{
		EcuID:               string(raw[13:25]),
		Version:             version,
		InvertersRegistered: int(binary.BigEndian.Uint16(raw[46:48])),
		InvertersOnline:     int(binary.BigEndian.Uint16(raw[48:50])),
		EthernetMac:         byteSliceToString(raw[55+verLength+3+tzLength : 55+verLength+3+tzLength+6]),
		WirelessMac:         byteSliceToString(raw[55+verLength+3+tzLength+6 : 55+verLength+3+tzLength+6+6]),
		LifetimeEnergy:      int(binary.BigEndian.Uint32(raw[27:31])) * 100,
		TodayEnergy:         int(binary.BigEndian.Uint32(raw[35:39])) * 10,
		LastPower:           int(binary.BigEndian.Uint32(raw[31:35])) * 1,
		Raw:                 raw,
	}, nil
}

type ArrayInfo struct {
	Timestamp time.Time
	Inverters []InverterInfo
	Raw       []byte
}

type InverterInfo struct {
	ID          string
	Online      bool
	Model       string
	Frequency   float64 // 0.1 Hz resolution
	Temperature int     // Celsius
	PowerA      int
	PowerB      int
	PowerC      int
	PowerD      int
	VoltageA    int
}

/*
NewArrayInfo parses the ArrayInfo response into a struct. It accepts a
IANA timezone (e.g. Europe/Amsterdam) to parse the timestamp from the
binary body. If left empty ("") it defaults to UTC

	# Explanation general header
	# ----------------------------------
	#  1- 3 APS
	#  4- 5 CommandGroup
	#  6- 8 Datastring Framelength
	#  9-12 command
	# 13-14 MatchStatus
	# 15-16 EcuModel
	# 17-18 number of inverters
	# 19-25 timestamp
*/
func NewArrayInfo(raw []byte, tz string) (ArrayInfo, error) {
	// Validation
	err := validateLength(raw)
	if err != nil {
		return ArrayInfo{Raw: raw}, err
	}

	// Set timezone for the timestamp returned by the ECU-R
	if tz == "" {
		tz = DefaultTz
	}

	// Parsing the header
	// - timestamp: 19-25
	timestamp, err := binToTimestamp(raw[19:26], tz)
	if err != nil {
		return ArrayInfo{Raw: raw}, err
	}

	// Parsing the inverters (only works for QS1s for now)
	numInverters := int(binary.BigEndian.Uint16(raw[17:19]))
	var inverters []InverterInfo
	start := 26
	length := 23
	for i := 0; i < numInverters; i++ {
		inverter, err := NewInverterInfo(raw[start+i*length : start+(i+1)*length])
		if err != nil {
			return ArrayInfo{Raw: raw}, fmt.Errorf("could not parse inverter %d from body: %w", i+1, err)
		}
		inverters = append(inverters, inverter)
	}

	return ArrayInfo{
		Timestamp: timestamp,
		Inverters: inverters,
		Raw:       raw,
	}, nil
}

/*
NewInverterInfo parses the inverterInfo response into a struct

# Record for each type of inverter
#------------------------
# 0-5  26-31 Inverter ID (UID)
# 6    32 0 or 1 Marks online status of inverter instance
# 7    33-34 "0" unknown
# 8    35 1=YC600 and 3=QS1. may be 2=YC1000
# 9-10 36-37 Frequency /10
# 11-12 38-39 Temperature Celsius (-100)
# 13-14 40-41 Power A Channel A on Inverter
# 15-16 42-43 Voltage A Channel A on Inverter
# 17-18 44-45 Power B Channel B on Inverter
# 19-20 46-47 Voltage B Channel B on Inverter POWER C on QS1
# 21-23 48-51 END\n or POWER D on QS1 till END
*/
func NewInverterInfo(raw []byte) (InverterInfo, error) {
	if len(raw)-1 < 22 {
		return InverterInfo{}, fmt.Errorf("body too short (<22 chars) to parse inverter: %w", ErrMalformedBody)
	}
	model := ""
	switch string(raw[8]) {
	case "1": // YC600 - not validated
		return InverterInfo{
			ID:          byteSliceToString(raw[0:6]),
			Online:      raw[6] != 0,
			Model:       "YC600",
			Frequency:   float64(binary.BigEndian.Uint16(raw[9:11])) / 10,
			Temperature: int(binary.BigEndian.Uint16(raw[11:13])) - 100,
			PowerA:      int(binary.BigEndian.Uint16(raw[13:15])),
			VoltageA:    int(binary.BigEndian.Uint16(raw[15:17])),
			PowerB:      int(binary.BigEndian.Uint16(raw[17:19])),
			PowerC:      0,
			PowerD:      0,
		}, nil
	case "2": // YC1000 - not validated
		return InverterInfo{
			ID:          byteSliceToString(raw[0:6]),
			Online:      raw[6] != 0,
			Model:       "YC1000",
			Frequency:   float64(binary.BigEndian.Uint16(raw[9:11])) / 10,
			Temperature: int(binary.BigEndian.Uint16(raw[11:13])) - 100,
			PowerA:      int(binary.BigEndian.Uint16(raw[13:15])),
			VoltageA:    int(binary.BigEndian.Uint16(raw[15:17])),
			PowerB:      int(binary.BigEndian.Uint16(raw[17:19])),
			PowerC:      int(binary.BigEndian.Uint16(raw[21:23])),
			PowerD:      int(binary.BigEndian.Uint16(raw[25:27])),
		}, nil
	case "3": // QS1
		return InverterInfo{
			ID:          byteSliceToString(raw[0:6]),
			Online:      raw[6] != 0,
			Model:       "QS1",
			Frequency:   float64(binary.BigEndian.Uint16(raw[9:11])) / 10,
			Temperature: int(binary.BigEndian.Uint16(raw[11:13])) - 100,
			PowerA:      int(binary.BigEndian.Uint16(raw[13:15])),
			VoltageA:    int(binary.BigEndian.Uint16(raw[15:17])),
			PowerB:      int(binary.BigEndian.Uint16(raw[17:19])),
			PowerC:      int(binary.BigEndian.Uint16(raw[19:21])),
			PowerD:      int(binary.BigEndian.Uint16(raw[21:23])),
		}, nil
	default:
		model = "Other"
	}

	return InverterInfo{
		ID:     byteSliceToString(raw[0:6]),
		Online: raw[6] != 0,
		Model:  model,
	}, ErrUnknownInverterType
}

type InverterSignalInfo struct {
	Status    int
	Inverters []InverterSignal
	Raw       []byte
}

type InverterSignal struct {
	ID     string
	Signal int
}

/*
# Explanation general header
# command APS1100280030[EDU-ID]END
# ----------------------------------
# 1- 3 APS
# 4- 5 ID for communication
# 6- 9 Datastring length
# 10-15 unkown (command/answer)
# 16-21 Inverter ID (UID)
# 22 Signal strength 0-255 inverter (4 groups)
# 23-25 END or continues to next "InverterID*" but always marks end of datastring
*/
func NewInverterSignalinfo(raw []byte) (InverterSignalInfo, error) {
	// Validation
	err := validateLength(raw)
	if err != nil {
		return InverterSignalInfo{Raw: raw}, err
	}

	// ECU level information
	statusValue := raw[13:15]
	status, err := strconv.Atoi(string(statusValue))
	if err != nil {
		return InverterSignalInfo{}, fmt.Errorf("could not parse status from body: %w", err)
	}
	res := InverterSignalInfo{
		Raw:    raw,
		Status: status,
	}

	// Per inverter signal information
	numInverters := (len(raw) - 19) / 7
	for i := 0; i < numInverters; i++ {
		inv := InverterSignal{
			ID:     byteSliceToString(raw[15+(i*7) : 15+6+(i*7)]),
			Signal: int(raw[15+6+(i*7)]),
		}
		res.Inverters = append(res.Inverters, inv)
	}

	return res, nil
}

// validateLength returns and error if the binary body length does
// not match the length indicated in the header of the body
func validateLength(body []byte) error {
	// Minimum length to contain a length indication
	if len(body) < 8 {
		return fmt.Errorf("body length less than 8 characters. Length was %d: %w", len(body), ErrMalformedBody)
	}

	// Finishes with 'END\n'
	match := []byte{69, 78, 68, 10}
	for i := 0; i < 4; i++ {
		if body[len(body)-4+i] != match[i] {
			return fmt.Errorf("body does not end with 'END\\n; got body (%v)': %w", body, ErrMalformedBody)
		}
	}

	resLength, err := strconv.Atoi(string(body[5:9]))
	if err != nil {
		return fmt.Errorf("failed to parse body lenght from header: %w", err)
	}

	if len(body)-1 != resLength {
		return fmt.Errorf("body length does not match length specified in header; expected %d, got %d: %w", resLength, len(body)-1, ErrMalformedBody)
	}

	return nil
}

// APS decided to encode timestamps in such a way that the Hex values
// read 'as if' they are decimals represent the timestamp. e.g., the
// year 2021 is encoded as 0x2021, which is acually int(8226).
// this function converts the hex values to a string, and then parses
// them as integers in order to generate the timestamp
func binToTimestamp(body []byte, tz string) (time.Time, error) {
	if len(body) != 7 {
		return time.Now(), ErrMalformedBody
	}
	year, err := strconv.Atoi(fmt.Sprintf("%X%X", body[0], body[1]))
	if err != nil {
		return time.Now(), err
	}

	month, err := strconv.Atoi(fmt.Sprintf("%X", body[2]))
	if err != nil {
		return time.Now(), err
	}

	day, err := strconv.Atoi(fmt.Sprintf("%X", body[3]))
	if err != nil {
		return time.Now(), err
	}

	hour, err := strconv.Atoi(fmt.Sprintf("%X", body[4]))
	if err != nil {
		return time.Now(), err
	}

	min, err := strconv.Atoi(fmt.Sprintf("%X", body[5]))
	if err != nil {
		return time.Now(), err
	}

	sec, err := strconv.Atoi(fmt.Sprintf("%X", body[6]))
	if err != nil {
		return time.Now(), err
	}
	loc, _ := time.LoadLocation(tz)
	return time.Date(year, time.Month(month), day, hour, min, sec, 0, loc), nil
}

func byteSliceToString(body []byte) string {
	res := ""
	for _, b := range body {
		res = fmt.Sprintf("%s%02X", res, b)
	}
	return res
}
