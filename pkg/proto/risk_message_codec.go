package proto

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/xinchentechnote/fin-proto-go/codec"
	risk_bin "github.com/xinchentechnote/fin-proto-go/risk-bin/messages"
)

// BinaryRiskMessageCodec risk proto message codec
type BinaryRiskMessageCodec struct{}

// Decode implements MessageCodec.
func (b *BinaryRiskMessageCodec) Decode(data []byte) (interface{}, codec.BinaryCodec, error) {
	var rcBinary risk_bin.RcBinary
	var buf bytes.Buffer
	buf.Write(data)
	err := rcBinary.Decode(&buf)
	if err != nil {
		return nil, nil, err
	}
	return rcBinary.MsgType, rcBinary.Body, nil
}

// Encode implements MessageCodec.
func (b *BinaryRiskMessageCodec) Encode(ext interface{}, message codec.BinaryCodec) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgType := ext.(uint32)
	rcBinary := &risk_bin.RcBinary{
		Version: 0,
		MsgType: msgType,
		Body:    message,
	}
	var buf bytes.Buffer
	err := rcBinary.Encode(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeJSONMap implements MessageCodec.
func (b *BinaryRiskMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
	msgType, err := strconv.Atoi(message["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", message["MsgType"].(string))
	}
	data, e := b.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return b.Encode(uint32(msgType), data)
}

// JSONToStruct implements MessageCodec.
func (b *BinaryRiskMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	msgType, err := strconv.Atoi(jsonMap["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", jsonMap["MsgType"].(string))
	}
	message, err := risk_bin.NewMessageByMsgType(uint32(msgType))
	if err != nil {
		return nil, err
	}
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
