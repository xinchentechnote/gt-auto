package codec

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SseBinFramer is a framer for the risk binary protocol.
type SseBinFramer struct{}

// ProtoName implements Framer.
func (r *SseBinFramer) ProtoName() string {
	return BinarySSE
}

// ReadFrame implements Framer.
func (r *SseBinFramer) ReadFrame(conn net.Conn) ([]byte, error) {
	head := make([]byte, 16)
	_, err := io.ReadFull(conn, head)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	//bodylen + checksum
	bodyLen := binary.BigEndian.Uint32(head[12:16]) + 4
	body := make([]byte, bodyLen)
	_, er := io.ReadFull(conn, body)
	if er != nil {
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}
	return append(head, body...), nil
}
