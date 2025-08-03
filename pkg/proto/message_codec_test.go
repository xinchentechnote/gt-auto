package proto_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/fin-proto-go/codec"
	risk_bin "github.com/xinchentechnote/fin-proto-go/risk-bin/messages"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
)

type dummyMessage struct {
	Content string
}

func (d *dummyMessage) MsgType() uint32 {
	return 999999
}

func (d *dummyMessage) Encode(buf *bytes.Buffer) error {
	err := codec.PutString[uint16](buf, d.Content)
	return err
}

func (d *dummyMessage) Decode(buf *bytes.Buffer) error {
	data, err := codec.GetString[uint16](buf)
	d.Content = data
	return err
}

func init() {
	risk_bin.RegistryRcBinaryMsgTypeFactory(999999, func() codec.BinaryCodec {
		return &dummyMessage{}
	})
}

func TestRiskMessageCodec_EncodeDecode(t *testing.T) {
	codec := &proto.BinaryRiskMessageCodec{}

	original := &dummyMessage{
		Content: "test-tlv-payload",
	}

	encoded, err := codec.Encode(uint32(999999), original)
	assert.NoError(t, err)
	assert.Greater(t, len(encoded), 8)

	_, decodedMsg, err := codec.Decode(encoded)
	assert.NoError(t, err)

	decoded, ok := decodedMsg.(*dummyMessage)
	assert.True(t, ok)
	assert.Equal(t, original.Content, decoded.Content)
}

func TestRiskMessageCodec_EncodeJsonMap(t *testing.T) {
	codec := &proto.BinaryRiskMessageCodec{}

	original := map[string]interface{}{
		"MsgType": "999999",
		"Content": 9527,
	}

	encoded, err := codec.EncodeJSONMap(original)
	assert.NoError(t, err)
	assert.Equal(t, len(encoded), 18)

	_, decodedMsg, err := codec.Decode(encoded)
	assert.NoError(t, err)

	decoded, ok := decodedMsg.(*dummyMessage)
	assert.True(t, ok)
	assert.Equal(t, "9527", decoded.Content)
}

func TestRiskMessageCodec_ConvertMapToStruct(t *testing.T) {
	original := map[string]interface{}{
		"UniqueOrderID": "1",
		"ClOrdID":       "2",
		"SecurityID":    "3",
		"Side":          "4",
		"Price":         "5",
		"OrderQty":      "6",
		"OrdType":       "7",
		"Account":       "8",
	}
	var msg risk_bin.NewOrder
	if err := proto.ConvertMapToStruct(original, &msg); err != nil {
		t.Fatalf("failed to convert map to struct: %v", err)
	}
	assert.Equal(t, "1", msg.UniqueOrderId)
}
