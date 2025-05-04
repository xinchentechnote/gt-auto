package proto_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
)

func TestLogon_EncodeDecode(t *testing.T) {
	original := &proto.Logon{
		SenderCompID:     "SENDER123",
		TargetCompID:     "TARGET456",
		HeartBtInt:       30,
		Password:         "pass123",
		DefaultApplVerID: 1,
	}

	data, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.Logon
	err = decoded.Decode(data)
	assert.NoError(t, err)

	assert.Equal(t, original.SenderCompID, decoded.SenderCompID)
	assert.Equal(t, original.TargetCompID, decoded.TargetCompID)
	assert.Equal(t, original.HeartBtInt, decoded.HeartBtInt)
	assert.Equal(t, original.Password, decoded.Password)
	assert.Equal(t, original.DefaultApplVerID, decoded.DefaultApplVerID)
}

func TestLogout_EncodeDecode(t *testing.T) {
	original := &proto.Logout{
		SessionStatus: 5,
		Text:          "Session expired due to timeout",
	}

	data, err := original.Encode()
	assert.NoError(t, err)

	var decoded proto.Logout
	err = decoded.Decode(data)
	assert.NoError(t, err)

	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Text, decoded.Text)
}

func TestLogout_DecodeInvalid(t *testing.T) {
	shortData := []byte{1} // invalid, too short
	var msg proto.Logout
	err := msg.Decode(shortData)
	assert.ErrorIs(t, err, proto.ErrInvalidPacket)
}

func TestHeartbeat_EncodeDecode(t *testing.T) {
	original := &proto.Heartbeat{}

	data, err := original.Encode()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(data))

	err = original.Decode(data)
	assert.NoError(t, err)
}
