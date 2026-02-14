package protocol

import (
	"bytes"
	"encoding/binary"
)

const (
	OP_DATA = iota + 1
	OP_CONTROL
)

type ElephantHeader struct {
	VersionAndOp uint8 // low 4 bits: version, high 4 bits: op
	Reserved     uint8
	SessionID    uint16
	DataLength   uint16
}

func NewElephantHeader(op uint8, sessionID, dataLength uint16) ElephantHeader {
	return ElephantHeader{
		VersionAndOp: (op << 4) | (2 & 0x0F), // version=2 in low 4 bits, op in high 4 bits
		SessionID:    sessionID,
		DataLength:   dataLength,
	}
}

func NewElephantControlHeader(dataLength uint16) ElephantHeader {
	return ElephantHeader{
		VersionAndOp: (OP_CONTROL << 4) | (2 & 0x0F), // version=2 in low 4 bits, op in high 4 bits
		SessionID:    0,                              // control messages have sessionID=0
		DataLength:   dataLength,
	}
}

func (h *ElephantHeader) Version() uint8 { return h.VersionAndOp & 0x0F }
func (h *ElephantHeader) Op() uint8      { return h.VersionAndOp >> 4 }

func (h *ElephantHeader) SetOp(op uint8) {
	h.VersionAndOp = (op << 4) | (h.VersionAndOp & 0x0F)
}

func (h *ElephantHeader) SetVersion(version uint8) {
	h.VersionAndOp = (h.VersionAndOp & 0xF0) | (version & 0x0F)
}

func (h *ElephantHeader) ToBuffer() *bytes.Buffer {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, h)
	return buf
}

func FromBytes(byteBuf []byte) (*ElephantHeader, error) {
	buf := bytes.NewReader(byteBuf)
	var header ElephantHeader
	err := binary.Read(buf, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}
	return &header, nil
}
