package proto

import "errors"

// ErrInvalidPacket is returned when a packet is invalid.
var ErrInvalidPacket = errors.New("invalid packet")

// MessageCodec is an interface for encoding and decoding messages.
type MessageCodec interface {
	// EncodeJSONMap encodes a map into a byte slice.
	EncodeJSONMap(map[string]interface{}) ([]byte, error)
	// Encode encodes a Message into a byte slice.
	// It returns the encoded byte slice and an error if any.
	Encode(Message) ([]byte, error)
	// Decode decodes a byte slice into a Message.
	// It returns the decoded Message and an error if any.
	// The byte slice must be of the correct length for the message type.
	Decode([]byte) (Message, error)
}

// Message is an interface that all message types must implement.
// It defines methods for getting the message type, encoding, and decoding.
type Message interface {
	// MsgType returns the message type as a uint32.
	MsgType() uint32
	// Encode encodes the message into a byte slice.
	// It returns the encoded byte slice and an error if any.
	Encode() ([]byte, error)
	// Decode decodes a byte slice into the message.
	// It returns an error if the byte slice is not of the correct length.
	Decode([]byte) error
}
