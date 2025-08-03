package codec

import "net"

// Framer interface defines methods for framing messages in a specific protocol.
type Framer interface {
	// ProtoName returns the name of the protocol.
	ProtoName() string
	// ReadFrame reads a frame from the connection and returns the raw byte slice.
	ReadFrame(conn net.Conn) ([]byte, error)
}
