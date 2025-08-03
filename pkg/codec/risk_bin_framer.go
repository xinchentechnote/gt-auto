package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// RiskBinFramer is a framer for the risk binary protocol.
type RiskBinFramer struct{}

// ProtoName implements Framer.
func (r *RiskBinFramer) ProtoName() string {
	return BinaryRisk
}

// ReadFrame implements Framer.
func (r *RiskBinFramer) ReadFrame(conn net.Conn) ([]byte, error) {
	head := make([]byte, 12)
	_, err := io.ReadFull(conn, head)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	bodyLen := binary.BigEndian.Uint32(head[8:12])
	body := make([]byte, bodyLen)
	_, er := io.ReadFull(conn, body)
	if er != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	return append(head, body...), nil
}
