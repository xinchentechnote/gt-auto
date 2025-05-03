package proto

type CancelRequest struct{}

// Decode implements Message.
func (c *CancelRequest) Decode([]byte) error {
	panic("unimplemented")
}

// Encode implements Message.
func (c *CancelRequest) Encode() ([]byte, error) {
	panic("unimplemented")
}

// MsgType implements Message.
func (c *CancelRequest) MsgType() string {
	panic("unimplemented")
}
