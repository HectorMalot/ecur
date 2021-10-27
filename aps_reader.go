package ecur

import (
	"fmt"
	"io"
	"strconv"
)

// ApsRead reads a full APSystems ECU-R response and returns the data as a byteslice.
// Internally, it reads the first 9 bytes (byte 6-9 indicates the length of the response,
// encoded as ascii) to determine the response length. It then reads the remaining response
// until the end of the indicated length
func ApsRead(source io.Reader) ([]byte, error) {
	// Get first 9 bytes (required to determine body length)
	body := make([]byte, 9)
	n, err := io.ReadAtLeast(source, body, 9)
	if err != nil {
		return body, fmt.Errorf("aps_reader: error reading from source: %w", err)
	}

	// Determine length
	expectedLength, err := strconv.Atoi(string(body[5:9]))
	if err != nil {
		return body, fmt.Errorf("failed to parse body length from header: %w", err)
	}

	body2 := make([]byte, expectedLength-n+1)
	n2, err := io.ReadAtLeast(source, body2, expectedLength-n+1)
	if err != nil {
		return body2, fmt.Errorf("aps_reader: error reading from source: %w", err)
	}

	if n+n2 != expectedLength+1 {
		return nil, fmt.Errorf("aps_reader: length of respones did not match expected length from header (got %d, expected %d): %w", n+n2, expectedLength, ErrMalformedBody)
	}

	body = append(body, body2...)
	if err = validateLength(body); err != nil {
		return body, fmt.Errorf("aps_reader: could not validate body: %w", err)
	}
	return body, nil
}
