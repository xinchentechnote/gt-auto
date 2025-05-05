package proto

import (
	"encoding/binary"
	"fmt"
)

func init() {
	//当前实现参考了深圳证券交易所的二进制协议，进行了简化，仅为了演示功能实现
	//将来将基于antlr4定义dsl描述完整的深交所二进制协议，并生成对应的代码
	RegisterMessage(1, func() Message { return &Logon{} })
	RegisterMessage(2, func() Message { return &Logout{} })
	RegisterMessage(3, func() Message { return &Heartbeat{} })
	RegisterMessage(4, func() Message { return &BusinessReject{} })
	RegisterMessage(100101, func() Message { return &NewOrder{} })
	RegisterMessage(200102, func() Message { return &ExecutionConfirm{} })
	RegisterMessage(200115, func() Message { return &ExecutionReport{} })
	RegisterMessage(190007, func() Message { return &CancelRequest{} })
	RegisterMessage(290008, func() Message { return &CancelReject{} })
}

// SzseMessageFactory is a factory function that creates a new message instance.
// It returns a Message interface, which is implemented by all message types.
type SzseMessageFactory func() Message

var registry = map[uint32]SzseMessageFactory{}

// RegisterMessage registers a message type with its factory function.
func RegisterMessage(msgType uint32, factory SzseMessageFactory) {
	registry[msgType] = factory
}

// BinarySzseMessageCodec is a codec for encoding and decoding messages.
// It implements the MessageCodec interface for the SZSE binary protocol.
type BinarySzseMessageCodec struct{}

// Encode a message into a byte slice and prepends the message type and length.
func (codec *BinarySzseMessageCodec) Encode(message Message) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgType := message.MsgType()

	data, err := message.Encode()
	if err != nil {
		return nil, err
	}

	length := len(data)
	b := make([]byte, 8+length)
	binary.BigEndian.PutUint32(b[0:4], uint32(msgType))
	binary.BigEndian.PutUint32(b[4:8], uint32(length))
	copy(b[8:], data)

	return b, nil
}

// Decode a byte slice into a message.
func (codec *BinarySzseMessageCodec) Decode(data []byte) (Message, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	msgTypeInt := binary.BigEndian.Uint32(data[0:4])
	length := binary.BigEndian.Uint32(data[4:8])
	if len(data) < int(8+length) {
		return nil, fmt.Errorf("data length mismatch")
	}

	body := data[8 : 8+length]

	factory, ok := registry[msgTypeInt]
	if !ok {
		return nil, fmt.Errorf("unknown MsgType: %d", msgTypeInt)
	}

	msg := factory()
	if err := msg.Decode(body); err != nil {
		return nil, err
	}

	return msg, nil
}

// BinarySseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE binary protocol.
type BinarySseMessageCodec struct{}

// Decode implements MessageCodec.
func (b *BinarySseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (b *BinarySseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}

// StepSzseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SZSE step protocol.
type StepSzseMessageCodec struct{}

// Decode implements MessageCodec.
func (s *StepSzseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (s *StepSzseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}

// StepSseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE step protocol.
type StepSseMessageCodec struct{}

// Decode implements MessageCodec.
func (s *StepSseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (s *StepSseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}
