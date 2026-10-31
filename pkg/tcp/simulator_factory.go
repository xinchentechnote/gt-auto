package tcp

import (
	"fmt"

	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
	gtcodec "github.com/xinchentechnote/gt-auto/pkg/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
)

// CreateSimulator creates a simulator based on the provided configuration.
func CreateSimulator[T fin_codec.BinaryCodec](config config.SimulatorConfig) (Simulator[T], error) {
	// fbsdk 模拟器自带会话与编解码（Rust 核心经 cgo），不走通用 framer/codec 装配
	if config.Type == "fbsdk" || config.Protocol == gtcodec.Fbsdk {
		if FbsdkSimulatorHook == nil {
			return nil, newFbsdkSimulatorMissing()
		}
		simAny, err := FbsdkSimulatorHook(config)
		if err != nil {
			return nil, err
		}
		sim, ok := simAny.(Simulator[T])
		if !ok {
			return nil, fmt.Errorf("fbsdk 模拟器仅支持 %T 类型参数", fin_codec.BinaryCodec(nil))
		}
		return sim, nil
	}
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
			queue:         make(chan receivedMessage, receiveQueueCapacity),
			Codec:         codec,
			Framer:        framer,
		}, nil
	case "tgw":
		tgw := &TgwSimulator[T]{
			ListenAddress: config.ListenAddress,
			queue:         make(chan receivedMessage, receiveQueueCapacity),
			Codec:         codec,
			Framer:        framer,
		}
		// SSE 交易所侧会话自动应答：网关登录 → PlatformState Open
		// （counter-gw 约定：收到 Open 即登录就绪；LoginTimeout 5s）
		if config.Protocol == gtcodec.BinarySSE {
			if frame := gtcodec.SsePlatformOpenFrame(); frame != nil {
				tgw.AutoReply = map[uint32][]byte{gtcodec.SseMsgTypeLogon: frame}
			}
		}
		return tgw, nil
	default:
		return nil, fmt.Errorf("unknown simulator type: %s", config.Type)
	}
}
