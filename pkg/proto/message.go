package proto

import (
	"errors"

	"github.com/xinchentechnote/fin-proto-go/codec"
)

// ErrInvalidPacket is returned when a packet is invalid.
var ErrInvalidPacket = errors.New("invalid packet")

// MessageCodec is an interface for encoding and decoding messages.
type MessageCodec interface {
	// EncodeJSONMap encodes a map into a byte slice.
	EncodeJSONMap(map[string]interface{}) ([]byte, error)
	// JSONToStruct converts a JSON-like map to a Message.
	JSONToStruct(map[string]interface{}) (codec.BinaryCodec, error)
	// Encode encodes a Message into a byte slice.
	// It returns the encoded byte slice and an error if any.
	Encode(interface{}, codec.BinaryCodec) ([]byte, error)
	// Decode decodes a byte slice into a Message.
	// It returns the decoded Message and an error if any.
	// The byte slice must be of the correct length for the message type.
	Decode([]byte) (interface{}, codec.BinaryCodec, error)
}
