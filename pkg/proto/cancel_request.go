package proto

import (
	"bytes"
	"errors"
)

type CancelRequest struct {
	ClOrdID     string // 10 bytes
	OrigClOrdID string // 10 bytes
	SecurityID  string // 8 bytes
	Side        byte
}

func (m *CancelRequest) MsgType() uint32 {
	return 190007
}

func (m *CancelRequest) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.ClOrdID, 10))
	buf.Write(padRight(m.OrigClOrdID, 10))
	buf.Write(padRight(m.SecurityID, 8))
	buf.WriteByte(m.Side)
	return buf.Bytes(), nil
}

func (m *CancelRequest) Decode(data []byte) error {
	if len(data) < 29 {
		return errors.New("invalid OrderCancelRequest packet")
	}
	m.ClOrdID = string(bytes.Trim(data[0:10], "\x00"))
	m.OrigClOrdID = string(bytes.Trim(data[10:20], "\x00"))
	m.SecurityID = string(bytes.Trim(data[20:28], "\x00"))
	m.Side = data[28]
	return nil
}
