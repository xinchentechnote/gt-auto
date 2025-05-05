package proto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
)

type dummyMessage struct {
	Content string
}

func (d *dummyMessage) MsgType() uint32 {
	return 999999
}

func (d *dummyMessage) Encode() ([]byte, error) {
	return []byte(d.Content), nil
}

func (d *dummyMessage) Decode(data []byte) error {
	d.Content = string(data)
	return nil
}

func init() {
	proto.RegisterMessage(999999, func() proto.Message {
		return &dummyMessage{}
	})
}

func TestSzseMessageCodec_EncodeDecode(t *testing.T) {
	codec := &proto.BinarySzseMessageCodec{}

	original := &dummyMessage{
		Content: "test-tlv-payload",
	}

	encoded, err := codec.Encode(original)
	assert.NoError(t, err)
	assert.Greater(t, len(encoded), 8)

	decodedMsg, err := codec.Decode(encoded)
	assert.NoError(t, err)

	decoded, ok := decodedMsg.(*dummyMessage)
	assert.True(t, ok)
	assert.Equal(t, original.Content, decoded.Content)
}

func TestSzseMessageCodec_EncodeJsonMap(t *testing.T) {
	codec := &proto.BinarySzseMessageCodec{}

	original := map[string]interface{}{
		"MsgType": "999999",
		"Content": 9527,
	}

	encoded, err := codec.EncodeJSONMap(original)
	assert.NoError(t, err)
	assert.Greater(t, len(encoded), 8)

	decodedMsg, err := codec.Decode(encoded)
	assert.NoError(t, err)

	decoded, ok := decodedMsg.(*dummyMessage)
	assert.True(t, ok)
	assert.Equal(t, "9527", decoded.Content)
}

func TestSzseMessageCodec_ConvertMapToStruct(t *testing.T) {
	original := map[string]interface{}{
		"MsgType": "999999",
		"Content": 9527,
	}
	var msg dummyMessage
	if err := proto.ConvertMapToStruct(original, &msg); err != nil {
		t.Fatalf("failed to convert map to struct: %v", err)
	}
	assert.Equal(t, "9527", msg.Content)
}
