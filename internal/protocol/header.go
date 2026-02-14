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

type Frame struct {
	ElephantHeader
	Payload *bytes.Buffer
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

func FromBytes(byteBuf []byte) *ElephantHeader {
	buf := bytes.NewReader(byteBuf)
	var header ElephantHeader
	_ = binary.Read(buf, binary.BigEndian, &header)
	return &header
}

func ExtractFrame(byteBuf *bytes.Buffer) *Frame {
	if byteBuf.Len() < 6 {
		return nil // Not enough data for header
	}

	headerBytes := byteBuf.Bytes()[:6]
	header := FromBytes(headerBytes)
	if byteBuf.Len() < int(6+header.DataLength) {
		return nil // Not enough data for payload
	}

	_ = byteBuf.Next(6) // Consume header bytes
	payloadBytes := byteBuf.Next(int(header.DataLength))
	// Copy payload to avoid holding reference to the large underlying buffer
	payloadCopy := make([]byte, len(payloadBytes))
	copy(payloadCopy, payloadBytes)
	return &Frame{
		ElephantHeader: *header,
		Payload:        bytes.NewBuffer(payloadCopy),
	}

}
