package proto

type MessageCodec interface {
	Encode(Message) ([]byte, error)
	Decode([]byte) (Message, error)
}

type Message interface {
	MsgType() string
	Encode() ([]byte, error)
	Decode([]byte) error
}
