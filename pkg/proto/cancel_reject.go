package proto

import (
	"bytes"
	"errors"
)

// CancelReject represents a cancel reject message.
type CancelReject struct {
	ClOrdID      string // 10 bytes
	OrigClOrdID  string // 10 bytes
	CxlRejReason byte
	RejectText   string // 40 bytes
}

// MsgType returns the message type for CancelReject.
func (m *CancelReject) MsgType() uint32 {
	return 290008
}

// Encode the CancelReject message into a byte slice.
func (m *CancelReject) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.ClOrdID, 10))
	buf.Write(padRight(m.OrigClOrdID, 10))
	buf.WriteByte(m.CxlRejReason)
	buf.Write(padRight(m.RejectText, 40))
	return buf.Bytes(), nil
}

// Decode the CancelReject message from a byte slice.
func (m *CancelReject) Decode(data []byte) error {
	if len(data) < 61 {
		return errors.New("invalid CancelReject packet")
	}
	m.ClOrdID = string(bytes.Trim(data[0:10], "\x00"))
	m.OrigClOrdID = string(bytes.Trim(data[10:20], "\x00"))
	m.CxlRejReason = data[20]
	m.RejectText = string(bytes.Trim(data[21:61], "\x00"))
	return nil
}
