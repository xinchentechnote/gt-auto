package codec

import (
	"bytes"
	"fmt"

	"github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	szse_bin "github.com/xinchentechnote/fin-proto-szse-bin-go/messages"
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
	msgType, err := msgTypeFromMap(message)
	if err != nil {
		return nil, err
	}
	data, e := codec.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return codec.Encode(msgType, data)
}

// JSONToStruct implements MessageCodec.
func (codec *BinarySzseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
	msgType, err := msgTypeFromMap(jsonMap)
	if err != nil {
		return nil, err
	}
	message, err := szse_bin.NewSzseBinaryMessageByMsgType(msgType)
	if err != nil {
		return nil, err
	}
	switch msg := message.(type) {
	case *szse_bin.NewOrder:
		if applID, ok := jsonMap["ApplID"].(string); ok {
			ext, err := szse_bin.NewNewOrderMessageByApplId(applID)
			if err == nil {
				msg.ApplExtend = ext
			}
		}
	case *szse_bin.ExecutionConfirm:
		if applID, ok := jsonMap["ApplID"].(string); ok {
			ext, err := szse_bin.NewExecutionConfirmMessageByApplId(applID)
			if err == nil {
				msg.ApplExtend = ext
			}
		}
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
