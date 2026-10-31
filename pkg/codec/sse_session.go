package codec

import (
	"bytes"
	"fmt"

	sse_bin "github.com/xinchentechnote/fin-proto-sse-bin-go/messages"
)

// SSE 会话级消息类型（fin-proto-sse-bin-go）。
const (
	// SseMsgTypeLogon 网关 → 交易所 登录
	SseMsgTypeLogon uint32 = 40
	// SseMsgTypePlatformState 交易所 → 网关 平台状态（登录就绪证据）
	SseMsgTypePlatformState uint32 = 209
)

// SsePlatformOpenFrame —— PlatformState{PlatformID=0, State=2(Open)} 完整帧
// （含 body_length 回填与 Checksum），用于交易所侧模拟器对网关登录的自动应答
// （柜台网关约定：收到 PlatformState Open 即登录就绪）。
func SsePlatformOpenFrame() []byte {
	sb := sse_bin.SseBinary{
		MsgType:   SseMsgTypePlatformState,
		MsgSeqNum: 1,
		Body:      &sse_bin.PlatformState{PlatformId: 0, PlatformState: 2},
	}
	var buf bytes.Buffer
	if err := sb.Encode(&buf); err != nil {
		return nil
	}
	return buf.Bytes()
}

// fbsdk 协议钩子：fbsdk 模拟器依赖 cgo（fbsdk-ffi cdylib），以 `-tags fbsdk`
// 构建时由适配包注入，默认构建保持纯 Go。未注入时工厂返回指引性错误。
var (
	// FbsdkCodecHook 构造 fbsdk 消息编解码器。
	FbsdkCodecHook func() (MessageCodec, error)
	// FbsdkFramerHook 构造 fbsdk 帧读取器（SDK 直连场景不读帧，可为 nil）。
	FbsdkFramerHook func() (Framer, error)
)

// fbsdkHooksMissing —— 统一的未启用错误。
func fbsdkHooksMissing() error {
	return fmt.Errorf("fbsdk 协议需要 cgo 构建：go build -tags fbsdk（并先构建 fbsdk-ffi cdylib）")
}
