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
	codec := &proto.SzseMessageCodec{}

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
