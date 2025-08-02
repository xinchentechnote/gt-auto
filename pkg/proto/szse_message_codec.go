package proto

import (
	"github.com/xinchentechnote/fin-proto-go/codec"
)

// BinarySzseMessageCodec is a codec for encoding and decoding messages.
// It implements the MessageCodec interface for the SZSE binary protocol.
type BinarySzseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (codec *BinarySzseMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
	panic("unimplemented")
}

// JSONToStruct implements MessageCodec.
func (codec *BinarySzseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	panic("unimplemented")
}

// Encode a message into a byte slice and prepends the message type and length.
func (codec *BinarySzseMessageCodec) Encode(ext interface{}, message codec.BinaryCodec) ([]byte, error) {
	panic("unimplemented")
}

// Decode a byte slice into a message.
func (codec *BinarySzseMessageCodec) Decode(data []byte) (interface{}, codec.BinaryCodec, error) {
	panic("unimplemented")
}
