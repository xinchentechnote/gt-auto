package proto

import (
	"bytes"
	"encoding/binary"
)

type Logon struct {
	SenderCompID     string // 固定 20 bytes
	TargetCompID     string // 固定 20 bytes
	HeartBtInt       int32  // 4 bytes
	Password         string // 固定 20 bytes
	DefaultApplVerID byte   // 1 byte
}

func (m *Logon) MsgType() string {
	return "1"
}

func (m *Logon) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.SenderCompID, 20))
	buf.Write(padRight(m.TargetCompID, 20))
	binary.Write(buf, binary.BigEndian, m.HeartBtInt)
	buf.Write(padRight(m.Password, 20))
	buf.WriteByte(m.DefaultApplVerID)
	return buf.Bytes(), nil
}

func (m *Logon) Decode(data []byte) error {
	buf := bytes.NewReader(data)
	sender := make([]byte, 20)
	target := make([]byte, 20)
	password := make([]byte, 20)
	version := make([]byte, 1) // ✅ 用来读取 DefaultApplVerID

	if _, err := buf.Read(sender); err != nil {
		return err
	}
	if _, err := buf.Read(target); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &m.HeartBtInt); err != nil {
		return err
	}
	if _, err := buf.Read(password); err != nil {
		return err
	}
	if _, err := buf.Read(version); err != nil {
		return err
	}
	m.DefaultApplVerID = version[0]

	m.SenderCompID = string(bytes.TrimRight(sender, "\x00"))
	m.TargetCompID = string(bytes.TrimRight(target, "\x00"))
	m.Password = string(bytes.TrimRight(password, "\x00"))
	return nil
}

type Logout struct {
	SessionStatus byte   // 1 byte
	Text          string // 固定 100 bytes
}

func (m *Logout) MsgType() string {
	return "2"
}

func (m *Logout) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(m.SessionStatus)
	buf.Write(padRight(m.Text, 100))
	return buf.Bytes(), nil
}

func (m *Logout) Decode(data []byte) error {
	if len(data) < 101 {
		return ErrInvalidPacket
	}
	m.SessionStatus = data[0]
	m.Text = string(bytes.TrimRight(data[1:], "\x00"))
	return nil
}

type Heartbeat struct{}

func (m *Heartbeat) MsgType() string {
	return "3"
}

func (m *Heartbeat) Encode() ([]byte, error) {
	return []byte{}, nil
}

func (m *Heartbeat) Decode(data []byte) error {
	return nil
}
