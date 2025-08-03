package codec

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xinchentechnote/fin-proto-go/codec"
	sse_bin "github.com/xinchentechnote/fin-proto-go/sse-bin/messages"
)

// BinarySseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE binary protocol.
type BinarySseMessageCodec struct{}

// ProtoName implements MessageCodec.
func (codec *BinarySseMessageCodec) ProtoName() string {
	return BinarySSE
}

// EncodeJSONMap implements MessageCodec.
func (codec *BinarySseMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
	msgType, err := strconv.Atoi(message["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", message["MsgType"].(string))
	}
	data, e := codec.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return codec.Encode(uint32(msgType), data)
}

// JSONToStruct implements MessageCodec.
func (codec *BinarySseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	msgType, err := strconv.Atoi(jsonMap["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", jsonMap["MsgType"].(string))
	}
	message, err := sse_bin.NewSseBinaryMessageByMsgType(uint32(msgType))
	if err != nil {
		return nil, err
	}
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Decode implements MessageCodec.
func (codec *BinarySseMessageCodec) Decode(data []byte) (interface{}, codec.BinaryCodec, error) {
	var szseBinary sse_bin.SseBinary
	var buf bytes.Buffer
	buf.Write(data)
	err := szseBinary.Decode(&buf)
	if err != nil {
		return nil, nil, err
	}
	return szseBinary.MsgType, szseBinary.Body, nil
}

// Encode implements MessageCodec.
func (codec *BinarySseMessageCodec) Encode(ext interface{}, message codec.BinaryCodec) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgType := ext.(uint32)
	szseBinary := &sse_bin.SseBinary{
		MsgType: msgType,
		Body:    message,
	}
	var buf bytes.Buffer
	err := szseBinary.Encode(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
