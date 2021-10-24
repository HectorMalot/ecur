package ecur

import (
	"errors"
)

var (
	ErrCouldNotConnect     = errors.New("could not connect to ECU-R")
	ErrNotConnected        = errors.New("not connected to ECU-R")
	ErrMalformedBody       = errors.New("binary body not as expected")
	ErrUnknownInverterType = errors.New("unknown inverter type")
)
