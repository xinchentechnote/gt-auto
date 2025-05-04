package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type BusinessReject struct {
	RefMsgType           string // 2 bytes
	BusinessRejectReason int32
	BusinessRejectText   string // 40 bytes
}

func (m *BusinessReject) MsgType() uint32 {
	return 4
}

func (m *BusinessReject) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.RefMsgType, 2))
	binary.Write(buf, binary.BigEndian, m.BusinessRejectReason)
	buf.Write(padRight(m.BusinessRejectText, 40))
	return buf.Bytes(), nil
}

func (m *BusinessReject) Decode(data []byte) error {
	if len(data) < 46 {
		return errors.New("invalid BusinessReject packet")
	}
	m.RefMsgType = string(bytes.Trim(data[0:2], "\x00"))
	m.BusinessRejectReason = int32(binary.BigEndian.Uint32(data[2:6]))
	m.BusinessRejectText = string(bytes.Trim(data[6:46], "\x00"))
	return nil
}
