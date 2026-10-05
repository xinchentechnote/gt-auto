package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/fin-proto-go/codec"
	risk_bin "github.com/xinchentechnote/fin-proto-go/risk-bin/messages"
)

type dummyMessage struct {
	Content string
}

func (d *dummyMessage) MsgType() uint32 {
	return 999999
}

func (d *dummyMessage) Encode(buf *bytes.Buffer) error {
	err := codec.WriteString[uint16](buf, d.Content)
	return err
}

func (d *dummyMessage) Decode(buf *bytes.Buffer) error {
	data, err := codec.ReadString[uint16](buf)
	d.Content = data
	return err
}

func init() {
	risk_bin.RegistryRcBinaryMsgTypeFactory(999999, func() codec.BinaryCodec {
		return &dummyMessage{}
	})
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
	if err := ConvertMapToStruct(original, &msg); err != nil {
		t.Fatalf("failed to convert map to struct: %v", err)
	}
	assert.Equal(t, "1", msg.UniqueOrderId)
}

// TestCodecInvalidMsgTypeReturnsError verifies malformed MsgType values yield
// errors instead of panics across all codecs.
func TestCodecInvalidMsgTypeReturnsError(t *testing.T) {
	codecs := map[string]MessageCodec{
		"risk": &BinaryRiskMessageCodec{},
		"szse": &BinarySzseMessageCodec{},
		"sse":  &BinarySseMessageCodec{},
	}
	inputs := []struct {
		name    string
		message map[string]interface{}
		wantErr string
	}{
		{"missing", map[string]interface{}{"ClOrdID": "1"}, "missing MsgType"},
		{"non-string", map[string]interface{}{"MsgType": 123}, "must be a string"},
		{"non-numeric", map[string]interface{}{"MsgType": "abc"}, "invalid MsgType"},
	}
	for cname, c := range codecs {
		for _, in := range inputs {
			t.Run(cname+"/"+in.name, func(t *testing.T) {
				_, err := c.EncodeJSONMap(in.message)
				assert.ErrorContains(t, err, in.wantErr)
				_, err = c.JSONToStruct(in.message)
				assert.ErrorContains(t, err, in.wantErr)
			})
		}
	}
}

// TestSzseCodecNewOrderWithoutApplID verifies a NewOrder map without ApplID
// does not panic; the ApplExtend extension is simply skipped.
func TestSzseCodecNewOrderWithoutApplID(t *testing.T) {
	codec := &BinarySzseMessageCodec{}
	_, err := codec.JSONToStruct(map[string]interface{}{
		"MsgType": "100101",
		"ClOrdID": "c0001",
	})
	assert.NoError(t, err)
}
