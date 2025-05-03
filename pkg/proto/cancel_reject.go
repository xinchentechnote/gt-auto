package proto

type CancelReject struct{}

// Decode implements Message.
func (c *CancelReject) Decode([]byte) error {
	panic("unimplemented")
}

// Encode implements Message.
func (c *CancelReject) Encode() ([]byte, error) {
	panic("unimplemented")
}

// MsgType implements Message.
func (c *CancelReject) MsgType() string {
	panic("unimplemented")
}
