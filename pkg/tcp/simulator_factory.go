package tcp

import (
	"fmt"

	"github.com/xinchentechnote/fin-proto-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
)

// CreateSimulator creates a simulator based on the provided configuration.
func CreateSimulator[T codec.BinaryCodec](config config.SimulatorConfig) (Simulator[T], error) {
	codec, err := proto.GetDefaultMessageCodecFactory().GetCodec(config.Protocol)
	if err != nil {
		return nil, err
	}

	switch config.Type {
	case "oms":
		return &OmsSimulator[T]{
			ServerAddress: config.ServerAddress,
			Codec:         codec,
		}, nil
	case "tgw":
		return &TgwSimulator[T]{
			ListenAddress: config.ListenAddress,
			Codec:         codec,
		}, nil
	default:
		return nil, fmt.Errorf("unknown simulator type: %s", config.Type)
	}
}
