package tcp

import (
	"fmt"

	fin_codec "github.com/xinchentechnote/fin-proto-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
)

// CreateSimulator creates a simulator based on the provided configuration.
func CreateSimulator[T fin_codec.BinaryCodec](config config.SimulatorConfig) (Simulator[T], error) {
	framer, err := codec.GetDefaultMessageCodecFactory().GetFramer(config.Protocol)
	if err != nil {
		return nil, err
	}
	codec, err := codec.GetDefaultMessageCodecFactory().GetCodec(config.Protocol)
	if err != nil {
		return nil, err
	}
	switch config.Type {
	case "oms":
		return &OmsSimulator[T]{
			ServerAddress: config.ServerAddress,
			Codec:         codec,
			Framer:        framer,
		}, nil
	case "tgw":
		return &TgwSimulator[T]{
			ListenAddress: config.ListenAddress,
			Codec:         codec,
			Framer:        framer,
		}, nil
	default:
		return nil, fmt.Errorf("unknown simulator type: %s", config.Type)
	}
}
