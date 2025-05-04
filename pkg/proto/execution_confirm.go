package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

// ExecutionConfirm represents an execution confirmation message.
type ExecutionConfirm struct {
	OrderID      string // 12 bytes
	ClOrdID      string // 10 bytes
	OrigClOrdID  string // 10 bytes
	ExecID       string // 12 bytes
	ExecType     byte   // 1 byte
	OrdStatus    byte   // 1 byte
	OrdRejReason int32  // 4 bytes
}

// MsgType returns the message type for ExecutionConfirm.
func (m *ExecutionConfirm) MsgType() uint32 {
	return 200102
}

// Encode the ExecutionConfirm message into a byte slice.
func (m *ExecutionConfirm) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.OrderID, 12))
	buf.Write(padRight(m.ClOrdID, 10))
	buf.Write(padRight(m.OrigClOrdID, 10))
	buf.Write(padRight(m.ExecID, 12))

	buf.WriteByte(m.ExecType)
	buf.WriteByte(m.OrdStatus)

	if err := binary.Write(buf, binary.BigEndian, m.OrdRejReason); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode the ExecutionConfirm message from a byte slice.
func (m *ExecutionConfirm) Decode(data []byte) error {
	if len(data) < 50 {
		return errors.New("invalid ExecutionConfirm packet length")
	}

	m.OrderID = string(bytes.TrimRight(data[0:12], "\x00"))
	m.ClOrdID = string(bytes.TrimRight(data[12:22], "\x00"))
	m.OrigClOrdID = string(bytes.TrimRight(data[22:32], "\x00"))
	m.ExecID = string(bytes.TrimRight(data[32:44], "\x00"))
	m.ExecType = data[44]
	m.OrdStatus = data[45]
	m.OrdRejReason = int32(binary.BigEndian.Uint32(data[46:50]))

	return nil
}
