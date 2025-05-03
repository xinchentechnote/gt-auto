package proto

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

func init() {
	RegisterMessage(100101, func() Message {
		return &NewOrder{}
	})
	RegisterMessage(200102, func() Message {
		return &ExecutionConfirm{}
	})
}

// 全局消息注册器
type SzseMessageFactory func() Message

var registry = map[uint32]SzseMessageFactory{}

func RegisterMessage(msgType uint32, factory SzseMessageFactory) {
	registry[msgType] = factory
}

type SzseMessageCodec struct{}

func (codec *SzseMessageCodec) Encode(message Message) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgTypeStr := message.MsgType()
	msgTypeInt, err := strconv.Atoi(msgTypeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MsgType (not int): %s", msgTypeStr)
	}

	data, err := message.Encode()
	if err != nil {
		return nil, err
	}

	length := len(data)
	b := make([]byte, 8+length)
	binary.BigEndian.PutUint32(b[0:4], uint32(msgTypeInt))
	binary.BigEndian.PutUint32(b[4:8], uint32(length))
	copy(b[8:], data)

	return b, nil
}

func (codec *SzseMessageCodec) Decode(data []byte) (Message, error) {
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
