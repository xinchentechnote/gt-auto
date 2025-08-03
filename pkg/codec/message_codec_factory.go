package codec

import (
	"errors"
	"sync"
)

// GatewayProtocol 表示支持的网关通信协议
type GatewayProtocol string

const (
	// UnknownProtocol 未知协议
	UnknownProtocol GatewayProtocol = "unknown"
	//BinaryRisk simple risk control proto
	BinaryRisk = "binary-risk"
	// BinarySZSE shenzhen stock exchange binary protocol
	BinarySZSE = "binary-szse"

	// BinarySSE shanghai stock exchange binary protocol
	BinarySSE = "binary-sse"

	// StepSZSE shenzhen stock exchange step protocol
	StepSZSE = "step-szse"
	// StepSSE shanghai stock exchange step protocol
	StepSSE = "step-sse"
)

var (
	instance MessageCodecFactory
	once     sync.Once
)

// GetDefaultMessageCodecFactory returns the default MessageCodecFactory instance.
func GetDefaultMessageCodecFactory() MessageCodecFactory {
	once.Do(func() {
		instance = &DefaultMessageCodecFactory{}
	})
	return instance
}

// MessageCodecFactory is an interface for creating message codecs based on the protocol.
type MessageCodecFactory interface {
	GetCodec(proto string) (MessageCodec, error)
	GetFramer(proto string) (Framer, error)
}

// DefaultMessageCodecFactory is the default implementation of MessageCodecFactory.
type DefaultMessageCodecFactory struct {
}

// GetCodec returns a MessageCodec based on the provided protocol string.
// It returns an error if the protocol is not supported.
// The protocol string should be one of the constants defined in this package.
// The supported protocols are:
// - "binary-szse"
// - "binary-sse"
// - "step-szse"
// - "step-sse"
func (f *DefaultMessageCodecFactory) GetCodec(proto string) (MessageCodec, error) {
	switch proto {
	case string(BinaryRisk):
		return &BinaryRiskMessageCodec{}, nil
	case string(BinarySZSE):
		return &BinarySzseMessageCodec{}, nil
	case string(BinarySSE):
		return &BinarySseMessageCodec{}, nil
	default:
		ErrUnsupportedProtocol := errors.New("unsupported protocol")
		return nil, ErrUnsupportedProtocol
	}
}

// GetFramer returns a Framer based on the provided protocol.
func (f *DefaultMessageCodecFactory) GetFramer(proto string) (Framer, error) {
	switch proto {
	case string(BinaryRisk):
		return &RiskBinFramer{}, nil
	default:
		ErrUnsupportedProtocol := errors.New("unsupported protocol")
		return nil, ErrUnsupportedProtocol
	}
}
