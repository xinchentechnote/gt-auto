package proto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
)

func padRightStr(s string, length int) string {
	for len(s) < length {
		s += "\x00"
	}
	return s
}

func TestNewOrder_EncodeDecode(t *testing.T) {
	original := &proto.NewOrder{
		ClOrdID:    "ORD1234567",
		Price:      100500,
		Qty:        200,
		SecurityID: "000001",
		MarketID:   "101",
		ApplID:     "010",
		Side:       1,
	}

	encoded, err := original.Encode()
	assert.NoError(t, err)
	assert.Len(t, encoded, 46)

	decoded := &proto.NewOrder{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, decoded)
}

func TestExecutionConfirm_EncodeDecode(t *testing.T) {
	original := &proto.ExecutionConfirm{
		OrderID:      "ORD123456789",
		ClOrdID:      "CL12345678",
		OrigClOrdID:  "ORIG123456",
		ExecID:       "EXEC12345678",
		ExecType:     1,
		OrdStatus:    2,
		OrdRejReason: 404,
	}

	encoded, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.ExecutionConfirm
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, padRightStr(original.OrderID, 12), decoded.OrderID)
	assert.Equal(t, padRightStr(original.ClOrdID, 10), decoded.ClOrdID)
	assert.Equal(t, padRightStr(original.OrigClOrdID, 10), decoded.OrigClOrdID)
	assert.Equal(t, padRightStr(original.ExecID, 12), decoded.ExecID)
	assert.Equal(t, original.ExecType, decoded.ExecType)
	assert.Equal(t, original.OrdStatus, decoded.OrdStatus)
	assert.Equal(t, original.OrdRejReason, decoded.OrdRejReason)
}

func TestExecutionReport_EncodeDecode(t *testing.T) {
	original := &proto.ExecutionReport{
		ClOrdID:   "CLIENT123",
		ExecID:    "EXEC456789",
		ExecType:  1,
		OrdStatus: 2,
		LastPx:    12345678,
		LastQty:   100,
	}
	encoded, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.ExecutionReport
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	assert.Equal(t, original, &decoded)
}

func TestCancelRequest_EncodeDecode(t *testing.T) {
	original := &proto.CancelRequest{
		ClOrdID:     "CANCEL001",
		OrigClOrdID: "ORIG0001",
		SecurityID:  "000001",
		Side:        1,
	}
	encoded, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.CancelRequest
	err = decoded.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, &decoded)
}

func TestCancelReject_EncodeDecode(t *testing.T) {
	original := &proto.CancelReject{
		ClOrdID:      "CL123",
		OrigClOrdID:  "OR456",
		CxlRejReason: 3,
		RejectText:   "Too late to cancel",
	}
	encoded, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.CancelReject
	err = decoded.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, &decoded)
}

func TestBusinessReject_EncodeDecode(t *testing.T) {
	original := &proto.BusinessReject{
		RefMsgType:           "01",
		BusinessRejectReason: 99,
		BusinessRejectText:   "Invalid ApplID",
	}
	encoded, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.BusinessReject
	err = decoded.Decode(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, &decoded)
}
