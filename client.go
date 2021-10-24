package ecur

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"
)

type Client struct {
	ip   string
	port string

	cooldown time.Duration
	conn     net.Conn

	ecuID string
}

func NewClient(ip string, port int) (*Client, error) {
	return &Client{
		ip:       ip,
		port:     strconv.Itoa(port),
		cooldown: time.Millisecond * time.Duration(25),
		conn:     nil,
		ecuID:    "",
	}, nil
}

func (c *Client) GetData() (ECUResponse, error) {
	err := c.Connect()
	if err != nil {
		log.Printf("Could not connect to ECU: %s", err)
		return ECUResponse{}, err
	}

	// Get ECU-R information
	ecuInfo, err := c.GetECUInfo()
	if err != nil {
		log.Printf("Could not get ECU information: %s", err)
		return ECUResponse{ECUInfo: ecuInfo}, err
	}
	c.ecuID = ecuInfo.EcuID

	// Wait between steps to not overload the ECU controller
	time.Sleep(c.cooldown)

	// Get Inverter information
	arrayInfo, err := c.GetInverterInfo()
	if err != nil {
		log.Printf("Could not get inverter information: %s", err)
		return ECUResponse{ECUInfo: ecuInfo, ArrayInfo: arrayInfo}, err
	}

	// Wait between steps to not overload the ECU controller
	time.Sleep(c.cooldown)

	// Get Inverter signal strength
	inverterSignal, err := c.GetInverterSignal()
	if err != nil {
		log.Printf("Could not get inverter signal strength information: %s", err)
		return ECUResponse{ECUInfo: ecuInfo, ArrayInfo: arrayInfo, InverterSignalInfo: inverterSignal}, err
	}

	return ECUResponse{
		ECUInfo:            ecuInfo,
		ArrayInfo:          arrayInfo,
		InverterSignalInfo: inverterSignal,
	}, nil
}

// connects with the ECU-R, but does not send further data
func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.ip+":"+c.port)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

// Close closes the connection to the ECU-R
// typically called after collecing all data
func (c *Client) Close() error {
	if c.conn == nil {
		return ErrNotConnected
	}
	return c.conn.Close()
}

// GetECUInfo() is the first call to the ECU and returns ECU information required
// for further communication with the ECU. Primarily the ECU-R ID
func (c *Client) GetECUInfo() (ECUInfo, error) {
	if c.conn == nil {
		return ECUInfo{}, ErrNotConnected
	}

	fmt.Fprint(c.conn, CmdECUInfo)
	raw, err := bufio.NewReader(c.conn).ReadBytes('\n')
	if err != nil {
		return ECUInfo{Raw: raw}, ErrMalformedBody
	}

	ecuInfo, err := NewECUInfo(raw)
	if err != nil {
		return ECUInfo{Raw: raw}, err
	}

	return ecuInfo, nil
}

// GetInverterInfo() is the second call tot he ECU and returns information
// per inverter, as well as information per MPPT for each inverter
func (c *Client) GetInverterInfo() (ArrayInfo, error) {
	if c.conn == nil {
		return ArrayInfo{}, ErrNotConnected
	}

	// Get ecuID if required
	if c.ecuID == "" {
		ecuInfo, err := c.GetECUInfo()
		if err != nil {
			return ArrayInfo{}, err
		}
		c.ecuID = ecuInfo.EcuID
	}

	// Run command
	fmt.Fprintf(c.conn, "%s%s%s", CmdInverterInfoPrefix, c.ecuID, CmdInverterInfoSuffix)
	raw, err := bufio.NewReader(c.conn).ReadBytes('\n')
	if err != nil {
		return ArrayInfo{}, ErrMalformedBody
	}

	arrayInfo, err := NewArrayInfo(raw)
	if err != nil {
		return ArrayInfo{}, err
	}

	return arrayInfo, nil
}

// GetInverterSignal() gets the Zigbee signal strength per inverter (scale 0x00-0xFF)
func (c *Client) GetInverterSignal() (InverterSignalInfo, error) {
	if c.conn == nil {
		return InverterSignalInfo{}, ErrNotConnected
	}

	// Get ecuID if required
	if c.ecuID == "" {
		ecuInfo, err := c.GetECUInfo()
		if err != nil {
			return InverterSignalInfo{}, err
		}
		c.ecuID = ecuInfo.EcuID
	}

	// Run command
	fmt.Fprintf(c.conn, "%s%s%s", CmdInverterSignalPrefix, c.ecuID, CmdInverterSignalSuffix)
	raw, err := bufio.NewReader(c.conn).ReadBytes('\n')
	if err != nil {
		return InverterSignalInfo{}, ErrMalformedBody
	}

	signalInfo, err := NewInverterSignalinfo(raw)
	if err != nil {
		return InverterSignalInfo{}, err
	}

	return signalInfo, nil
}