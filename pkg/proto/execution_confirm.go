package proto

type ExecutionConfirm struct {
	OrderID      string //交易所订单编号
	ClOrdID      string //客户订单编号
	OrigClOrdID  string //原始订单客户订单编号
	ExecID       string //执行编号
	ExecType     string //执行类型
	OrdStatus    string //订单状态
	OrdRejReason string //撤单/拒绝原因代码
}

func (m *ExecutionConfirm) MsgType() string {
	return "200102"
}

func (m *ExecutionConfirm) Encode() ([]byte, error) {
	// TODO
	return []byte{}, nil
}

func (m *ExecutionConfirm) Decode(data []byte) error {
	// TODO
	return nil
}
