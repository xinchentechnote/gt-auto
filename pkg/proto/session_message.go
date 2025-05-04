package proto

import (
	"bytes"
	"encoding/binary"
)

// Logon represents a logon message.
type Logon struct {
	SenderCompID     string // fixed 20 bytes
	TargetCompID     string // fixed 20 bytes
	HeartBtInt       int32  // 4 bytes
	Password         string // fixed 20 bytes
	DefaultApplVerID byte   // 1 byte
}

// MsgType returns the message type for Logon.
func (m *Logon) MsgType() uint32 {
	return 1
}

// Encode the Logon message into a byte slice.
func (m *Logon) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(padRight(m.SenderCompID, 20))
	buf.Write(padRight(m.TargetCompID, 20))
	if err := binary.Write(buf, binary.BigEndian, m.HeartBtInt); err != nil {
		return nil, err
	}
	buf.Write(padRight(m.Password, 20))
	buf.WriteByte(m.DefaultApplVerID)
	return buf.Bytes(), nil
}

// Decode the Logon message from a byte slice.
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

// Logout represents a logout message.
type Logout struct {
	SessionStatus byte   // 1 byte
	Text          string // 固定 100 bytes
}

// MsgType returns the message type for Logout.
func (m *Logout) MsgType() uint32 {
	return 2
}

// Encode the Logout message into a byte slice.
func (m *Logout) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(m.SessionStatus)
	buf.Write(padRight(m.Text, 100))
	return buf.Bytes(), nil
}

// Decode the Logout message from a byte slice.
func (m *Logout) Decode(data []byte) error {
	if len(data) < 101 {
		return ErrInvalidPacket
	}
	m.SessionStatus = data[0]
	m.Text = string(bytes.TrimRight(data[1:], "\x00"))
	return nil
}

// Heartbeat represents a heartbeat message.
type Heartbeat struct{}

// MsgType returns the message type for Heartbeat.
func (m *Heartbeat) MsgType() uint32 {
	return 3
}

// Encode the Heartbeat message into a byte slice.
func (m *Heartbeat) Encode() ([]byte, error) {
	return []byte{}, nil
}

// Decode the Heartbeat message from a byte slice.
func (m *Heartbeat) Decode(data []byte) error {
	return nil
}
