package codec

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xinchentechnote/fin-proto-go/codec"
	szse_bin "github.com/xinchentechnote/fin-proto-go/szse-bin/messages"
)

// BinarySzseMessageCodec is a codec for encoding and decoding messages.
// It implements the MessageCodec interface for the SZSE binary protocol.
type BinarySzseMessageCodec struct{}

// ProtoName implements MessageCodec.
func (codec *BinarySzseMessageCodec) ProtoName() string {
	return BinarySZSE
}

// EncodeJSONMap implements MessageCodec.
func (codec *BinarySzseMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
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
func (codec *BinarySzseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	msgType, err := strconv.Atoi(jsonMap["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", jsonMap["MsgType"].(string))
	}
	message, err := szse_bin.NewSzseBinaryMessageByMsgType(uint32(msgType))
	if err != nil {
		return nil, err
	}
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Encode a message into a byte slice and prepends the message type and length.
func (codec *BinarySzseMessageCodec) Encode(ext interface{}, message codec.BinaryCodec) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgType := ext.(uint32)
	szseBinary := &szse_bin.SzseBinary{
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

// Decode a byte slice into a message.
func (codec *BinarySzseMessageCodec) Decode(data []byte) (interface{}, codec.BinaryCodec, error) {
	var szseBinary szse_bin.SzseBinary
	var buf bytes.Buffer
	buf.Write(data)
	err := szseBinary.Decode(&buf)
	if err != nil {
		return nil, nil, err
	}
	return szseBinary.MsgType, szseBinary.Body, nil
}
