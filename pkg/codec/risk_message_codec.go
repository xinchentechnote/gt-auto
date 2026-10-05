package codec

import (
	"bytes"
	"fmt"

	risk_bin "github.com/xinchentechnote/fin-proto-risk-bin-go/messages"
	"github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
)

// BinaryRiskMessageCodec risk proto message codec
type BinaryRiskMessageCodec struct{}

// ProtoName implements MessageCodec.
func (b *BinaryRiskMessageCodec) ProtoName() string {
	return BinaryRisk
}

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

// Encode encodes a message into a byte slice prefixed with its message type.
func (b *BinaryRiskMessageCodec) Encode(msgType uint32, message codec.BinaryCodec) ([]byte, error) {
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
	msgType, err := msgTypeFromMap(message)
	if err != nil {
		return nil, err
	}
	data, e := b.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return b.Encode(msgType, data)
}

// JSONToStruct implements MessageCodec.
func (b *BinaryRiskMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	msgType, err := msgTypeFromMap(jsonMap)
	if err != nil {
		return nil, err
	}
	message, err := risk_bin.NewRcBinaryMessageByMsgType(msgType)
	if err != nil {
		return nil, err
	}
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}
