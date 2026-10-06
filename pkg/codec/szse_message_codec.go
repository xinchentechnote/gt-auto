package codec

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	szse_bin "github.com/xinchentechnote/fin-proto-szse-bin-go/messages"
)

// BinarySzseMessageCodec is a codec for encoding and decoding messages.
// It implements the MessageCodec interface for the SZSE binary protocol.
type BinarySzseMessageCodec struct{}

// ProtoName implements MessageCodec.
func (c *BinarySzseMessageCodec) ProtoName() string {
	return BinarySZSE
}

// EncodeJSONMap implements MessageCodec.
func (c *BinarySzseMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
	msgType, err := msgTypeFromMap(message)
	if err != nil {
		return nil, err
	}
	data, e := c.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return c.Encode(msgType, data)
}

// JSONToStruct implements MessageCodec.
func (c *BinarySzseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (codec.BinaryCodec, error) {
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
		msg.ApplExtend = applExtendFor(jsonMap, msg, szse_bin.NewNewOrderMessageByApplId)
	case *szse_bin.ExecutionConfirm:
		msg.ApplExtend = applExtendFor(jsonMap, msg, szse_bin.NewExecutionConfirmMessageByApplId)
	case *szse_bin.ExecutionReport:
		msg.ApplExtend = applExtendFor(jsonMap, msg, szse_bin.NewExecutionReportMessageByApplId)
	}
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// applExtendFor 按 ApplID 预填消息的扩展字段结构。必须覆盖所有带 ApplExtend 的消息类型：
// 协议解码端会按 ApplId 无条件实例化扩展（非 nil），期望消息若留 nil 会导致
// Receive 步骤的全字段对比出现 nil≠空结构 的误报。
func applExtendFor[F codec.BinaryCodec](jsonMap map[string]interface{}, msg codec.BinaryCodec, factory func(string) (F, error)) codec.BinaryCodec {
	applID, ok := jsonMap["ApplID"].(string)
	if !ok {
		log.Warnf("%T data has no ApplID, skipping ApplExtend", msg)
		return nil
	}
	ext, err := factory(applID)
	if err != nil {
		log.Warnf("ApplID %q not registered for %T, skipping extension: %v", applID, msg, err)
		return nil
	}
	return ext
}

// Encode encodes a message into a byte slice prefixed with its message type.
func (c *BinarySzseMessageCodec) Encode(msgType uint32, message codec.BinaryCodec) ([]byte, error) {
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
func (c *BinarySzseMessageCodec) Decode(data []byte) (interface{}, codec.BinaryCodec, error) {
	var szseBinary szse_bin.SzseBinary
	var buf bytes.Buffer
	buf.Write(data)
	err := szseBinary.Decode(&buf)
	if err != nil {
		return nil, nil, err
	}
	return szseBinary.MsgType, szseBinary.Body, nil
}
