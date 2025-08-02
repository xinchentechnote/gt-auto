package proto

import "github.com/xinchentechnote/fin-proto-go/codec"

// BinarySseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE binary protocol.
type BinarySseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (b *BinarySseMessageCodec) EncodeJSONMap(map[string]interface{}) ([]byte, error) {
	panic("unimplemented")
}

// JSONToStruct implements MessageCodec.
func (b *BinarySseMessageCodec) JSONToStruct(map[string]interface{}) (codec.BinaryCodec, error) {
	panic("unimplemented")
}

// Decode implements MessageCodec.
func (b *BinarySseMessageCodec) Decode([]byte) (interface{}, codec.BinaryCodec, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (b *BinarySseMessageCodec) Encode(interface{}, codec.BinaryCodec) ([]byte, error) {
	panic("unimplemented")
}
