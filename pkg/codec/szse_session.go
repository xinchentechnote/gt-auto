package codec

import (
	"bytes"

	szse_bin "github.com/xinchentechnote/fin-proto-szse-bin-go/messages"
)

// SZSE 会话级消息类型（fin-proto-szse-bin-go）。
const (
	// SzseMsgTypeLogon 网关 → 交易所 登录
	SzseMsgTypeLogon uint32 = 1
	// SzseMsgTypePlatformState 交易所 → 网关 平台状态（登录就绪证据）
	SzseMsgTypePlatformState uint32 = 6
)

// SzsePlatformOpenFrame —— PlatformStateInfo{PlatformID=1(现货集中竞价),
// State=2(Open)} 完整帧（含 body_length 回填与 SZSE_BIN 校验和），用于交易所侧
// 模拟器对网关登录的自动应答（counter-gw-szse 登录就绪约定）。
func SzsePlatformOpenFrame() []byte {
	sb := szse_bin.SzseBinary{
		MsgType: SzseMsgTypePlatformState,
		Body:    &szse_bin.PlatformStateInfo{PlatformId: 1, PlatformState: 2},
	}
	var buf bytes.Buffer
	if err := sb.Encode(&buf); err != nil {
		return nil
	}
	return buf.Bytes()
}
