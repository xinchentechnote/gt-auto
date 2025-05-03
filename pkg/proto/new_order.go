package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)

type NewOrder struct {
	ClOrdID    string // fixed 10 bytes
	Price      int64  // 8 bytes
	Qty        int32  // 4 bytes
	SecurityID string // fixed 12 bytes
	MarketID   string // fixed 8 bytes
	ApplID     string // fixed 3 bytes
	Side       byte   // 1 byte
}

func (m *NewOrder) MsgType() string {
	return "100101"
}

func padRight(s string, length int) []byte {
	b := make([]byte, length)
	copy(b, []byte(s))
	return b
}

func (m *NewOrder) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Write ClOrdID (10 bytes)
	buf.Write(padRight(m.ClOrdID, 10))

	// Write Price (8 bytes)
	if err := binary.Write(buf, binary.BigEndian, m.Price); err != nil {
		return nil, err
	}

	// Write Qty (4 bytes)
	if err := binary.Write(buf, binary.BigEndian, m.Qty); err != nil {
		return nil, err
	}

	// Write SecurityID (12 bytes)
	buf.Write(padRight(m.SecurityID, 12))
	// Write MarketID (8 bytes)
	buf.Write(padRight(m.MarketID, 8))
	// Write ApplID (3 bytes)
	buf.Write(padRight(m.ApplID, 3))

	// Write Side (1 byte)
	buf.WriteByte(m.Side)

	return buf.Bytes(), nil
}

func (m *NewOrder) Decode(data []byte) error {
	if len(data) < 46 { // 10 + 8 + 4 + 12 + 8 + 3 + 1
		return errors.New("invalid NewOrder packet length")
	}

	buf := bytes.NewReader(data)

	// Read ClOrdID
	clOrdID := make([]byte, 10)
	if _, err := buf.Read(clOrdID); err != nil {
		return err
	}
	m.ClOrdID = string(bytes.TrimRight(clOrdID, "\x00"))

	// Read Price
	if err := binary.Read(buf, binary.BigEndian, &m.Price); err != nil {
		return err
	}

	// Read Qty
	if err := binary.Read(buf, binary.BigEndian, &m.Qty); err != nil {
		return err
	}

	// Read SecurityID
	securityID := make([]byte, 12)
	if _, err := buf.Read(securityID); err != nil {
		return err
	}
	m.SecurityID = string(bytes.TrimRight(securityID, "\x00"))
	// Read MarketID
	marketID := make([]byte, 8)
	if _, err := buf.Read(marketID); err != nil {
		return err
	}
	m.MarketID = string(bytes.TrimRight(marketID, "\x00"))
	// Read ApplID
	applID := make([]byte, 3)
	if _, err := buf.Read(applID); err != nil {
		return err
	}
	m.ApplID = string(bytes.TrimRight(applID, "\x00"))

	// Read Side
	if err := binary.Read(buf, binary.BigEndian, &m.Side); err != nil {
		return err
	}

	return nil
}
