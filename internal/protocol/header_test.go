package protocol

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestHeaderSize(t *testing.T) {
	header := NewElephantHeader(1, 1234, 5678)
	assert.Equal(t, uintptr(6), unsafe.Sizeof(header))
}

func TestBigEndianEncoding(t *testing.T) {
	header := NewElephantHeader(2, 1, 65533)
	expectedBytes := []byte{0x22, 0x00, 0x00, 0x01, 0xff, 0xfd}
	assert.Equal(t, expectedBytes, header.ToBuffer().Bytes())

	newHeader, err := FromBytes(expectedBytes)
	assert.NoError(t, err)
	assert.Equal(t, header, *newHeader)

}

func TestHeaderOPCodes(t *testing.T) {
	assert.Equal(t, 1, OP_DATA)
	assert.Equal(t, 2, OP_CONTROL)
}

func TestNewElephantControlHeader(t *testing.T) {
	header := NewElephantControlHeader(100)
	assert.Equal(t, uint8(0x22), header.VersionAndOp)
	assert.Equal(t, uint16(0), header.SessionID)
	assert.Equal(t, uint16(100), header.DataLength)
}
