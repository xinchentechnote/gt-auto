package proto

type ExecutionReport struct{}

// Decode implements Message.
func (e *ExecutionReport) Decode([]byte) error {
	panic("unimplemented")
}

// Encode implements Message.
func (e *ExecutionReport) Encode() ([]byte, error) {
	panic("unimplemented")
}

// MsgType implements Message.
func (e *ExecutionReport) MsgType() string {
	panic("unimplemented")
}
