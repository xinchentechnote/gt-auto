package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SzseBinFramer is a framer for the risk binary protocol.
type SzseBinFramer struct{}

// ProtoName implements Framer.
func (r *SzseBinFramer) ProtoName() string {
	return BinarySZSE
}

// ReadFrame implements Framer.
func (r *SzseBinFramer) ReadFrame(conn net.Conn) ([]byte, error) {
	head := make([]byte, 8)
	_, err := io.ReadFull(conn, head)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	//bodylen + 4
	bodyLen := binary.BigEndian.Uint32(head[4:8]) + 4
	body := make([]byte, bodyLen)
	_, er := io.ReadFull(conn, body)
	if er != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	return append(head, body...), nil
}
