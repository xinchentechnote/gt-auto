package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// ExecutionReport represents an execution report message.
type ExecutionReport struct {
	ClOrdID   string // 10 bytes
	ExecID    string // 12 bytes
	ExecType  byte
	OrdStatus byte
	LastPx    int64 // 8 bytes
	LastQty   int32 // 4 bytes
}

// MsgType returns the message type for ExecutionReport.
func (m *ExecutionReport) MsgType() uint32 {
	return 200115
}

// Encode the ExecutionReport message into a byte slice.
func (m *ExecutionReport) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.ClOrdID, 10))
	buf.Write(padRight(m.ExecID, 12))
	buf.WriteByte(m.ExecType)
	buf.WriteByte(m.OrdStatus)
	if err := binary.Write(buf, binary.BigEndian, m.LastPx); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, m.LastQty); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode the ExecutionReport message from a byte slice.
func (m *ExecutionReport) Decode(data []byte) error {
	if len(data) < 36 {
		return errors.New("invalid ExecutionReport packet")
	}
	m.ClOrdID = string(bytes.Trim(data[0:10], "\x00"))
	m.ExecID = string(bytes.Trim(data[10:22], "\x00"))
	m.ExecType = data[22]
	m.OrdStatus = data[23]
	m.LastPx = int64(binary.BigEndian.Uint64(data[24:32]))
	m.LastQty = int32(binary.BigEndian.Uint32(data[32:36]))
	return nil
}
