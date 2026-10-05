package codec

import (
	"fmt"
)

const (
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

// defaultMessageCodecFactory is the shared factory instance; the factory is
// stateless, so a lazily initialized singleton would add nothing.
var defaultMessageCodecFactory MessageCodecFactory = &DefaultMessageCodecFactory{}

// GetDefaultMessageCodecFactory returns the default MessageCodecFactory instance.
func GetDefaultMessageCodecFactory() MessageCodecFactory {
	return defaultMessageCodecFactory
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
// The supported protocols are:
// - "binary-risk"
// - "binary-szse"
// - "binary-sse"
func (f *DefaultMessageCodecFactory) GetCodec(proto string) (MessageCodec, error) {
	switch proto {
	case BinaryRisk:
		return &BinaryRiskMessageCodec{}, nil
	case BinarySZSE:
		return &BinarySzseMessageCodec{}, nil
	case BinarySSE:
		return &BinarySseMessageCodec{}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}

// GetFramer returns a Framer based on the provided protocol.
func (f *DefaultMessageCodecFactory) GetFramer(proto string) (Framer, error) {
	switch proto {
	case BinaryRisk:
		return &RiskBinFramer{}, nil
	case BinarySZSE:
		return &SzseBinFramer{}, nil
	case BinarySSE:
		return &SseBinFramer{}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", proto)
	}
}
