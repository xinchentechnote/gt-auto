package proto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrder_EncodeDecode(t *testing.T) {
	original := &NewOrder{
		ClOrdID:    "ORD1234567",
		Price:      100500,
		Qty:        200,
		SecurityID: "000001",
		MarketID:   "101",
		ApplID:     "010",
		Side:       1,
	}

	// Encode to bytes
	encoded, err := original.Encode()
	assert.NoError(t, err)
	assert.Len(t, encoded, 46, "encoded NewOrder should be exactly 46 bytes")

	// Decode back
	decoded := &NewOrder{}
	err = decoded.Decode(encoded)
	assert.NoError(t, err)

	// Check equality
	assert.Equal(t, original.ClOrdID, decoded.ClOrdID)
	assert.Equal(t, original.Price, decoded.Price)
	assert.Equal(t, original.Qty, decoded.Qty)
	assert.Equal(t, original.SecurityID, decoded.SecurityID)
	assert.Equal(t, original.MarketID, decoded.MarketID)
	assert.Equal(t, original.ApplID, decoded.ApplID)
	assert.Equal(t, original.Side, decoded.Side)
}
