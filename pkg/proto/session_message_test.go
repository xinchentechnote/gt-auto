package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogonEncodeDecode(t *testing.T) {
	original := &Logon{
		SenderCompID:     "SENDER123",
		TargetCompID:     "TARGET456",
		HeartBtInt:       30,
		Password:         "pass123",
		DefaultApplVerID: 1,
	}

	data, err := original.Encode()
	assert.NoError(t, err)

	var decoded Logon
	err = decoded.Decode(data)
	assert.NoError(t, err)

	assert.Equal(t, original.SenderCompID, decoded.SenderCompID)
	assert.Equal(t, original.TargetCompID, decoded.TargetCompID)
	assert.Equal(t, original.HeartBtInt, decoded.HeartBtInt)
	assert.Equal(t, original.Password, decoded.Password)
	assert.Equal(t, original.DefaultApplVerID, decoded.DefaultApplVerID)
}

func TestLogoutEncodeDecode(t *testing.T) {
	original := &Logout{
		SessionStatus: 5,
		Text:          "Session expired due to timeout",
	}

	data, err := original.Encode()
	assert.NoError(t, err)

	var decoded Logout
	err = decoded.Decode(data)
	assert.NoError(t, err)

	assert.Equal(t, original.SessionStatus, decoded.SessionStatus)
	assert.Equal(t, original.Text, decoded.Text)
}

func TestLogoutDecodeInvalid(t *testing.T) {
	shortData := []byte{1} // invalid, too short
	var msg Logout
	err := msg.Decode(shortData)
	assert.ErrorIs(t, err, ErrInvalidPacket)
}

func TestHeartbeatEncodeDecode(t *testing.T) {
	original := &Heartbeat{}

	data, err := original.Encode()
	assert.NoError(t, err)
	assert.Equal(t, 0, len(data))

	err = original.Decode(data)
	assert.NoError(t, err)
}
