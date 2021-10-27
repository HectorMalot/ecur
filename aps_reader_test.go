package ecur

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApsReader(t *testing.T) {
	// ArrayInfo body with a byte(10) somewhere in the middle (which would trip a search for \n)
	input := []byte{65, 80, 83, 49, 49, 48, 48, 55, 53, 48, 48, 48, 50, 48, 48, 48, 49, 0, 2, 32, 33, 16, 32, 20, 24, 5, 128, 16, 0, 3, 0, 0, 1, 48, 51, 1, 243, 0, 119, 0, 57, 0, 228, 0, 56, 0, 60, 0, 10, 128, 16, 0, 3, 0, 1, 1, 48, 51, 1, 243, 0, 118, 0, 55, 0, 229, 0, 55, 0, 57, 0, 56, 69, 78, 68, 10}
	buf := bytes.NewBuffer(input)

	body, err := ApsRead(buf)
	require.NoError(t, err)
	require.Equal(t, len(input), len(body))

}
