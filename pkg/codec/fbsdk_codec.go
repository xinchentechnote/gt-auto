//go:build fbsdk

// fbsdk 协议适配（-tags fbsdk 构建）：FB-SDK 柜台客户端模拟器的消息面。
//
// 字段结构体为**用例数据面**（CSV/JSON 表头 = 字段名，经 ConvertMapToStruct
// 宽松转换 + go-cmp 字段级比对）；线格式编解码由 Rust 核心（fbsdk-ffi）承担，
// Go 侧不做字节级处理。
package codec

import (
	"bytes"
	"errors"
	"fmt"

	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
)

// FCP 消息类型（fcp-proto；十进制，用例 msg_type 列使用）。
const (
	// FbsdkMsgNewOrder 新委托申报（SDK → 柜台）
	FbsdkMsgNewOrder uint32 = 64
	// FbsdkMsgCancel 撤单申报
	FbsdkMsgCancel uint32 = 65
	// FbsdkMsgOrderConfirm 委托确认（受理/撤单确认/拒绝三合一）
	FbsdkMsgOrderConfirm uint32 = 66
	// FbsdkMsgCancelReject 撤单失败
	FbsdkMsgCancelReject uint32 = 67
	// FbsdkMsgTrade 成交回报
	FbsdkMsgTrade uint32 = 68
	// FbsdkMsgQueryAccount 资金查询
	FbsdkMsgQueryAccount uint32 = 80
	// FbsdkMsgQueryPosition 持仓查询
	FbsdkMsgQueryPosition uint32 = 82
	// FbsdkMsgQueryOrder 当日委托查询
	FbsdkMsgQueryOrder uint32 = 84
	// FbsdkMsgQueryTrade 当日成交查询
	FbsdkMsgQueryTrade uint32 = 86
)

// FbsdkNewOrder —— 委托申报字段（用例数据面；线格式由 SDK 核心编码）。
type FbsdkNewOrder struct {
	ClOrdID       uint64
	SecurityID    uint32
	AccountID     uint32
	PriceMantissa int64
	PriceScale    int8
	OrderQty      int64
	Market        uint8
	Side          uint8
	OrdType       uint8
}

// FbsdkCancel —— 撤单申报字段。
type FbsdkCancel struct {
	ClOrdID     uint64
	OrigClOrdID uint64
	SecurityID  uint32
	AccountID   uint32
	Market      uint8
	Side        uint8
}

// FbsdkConfirm —— 委托确认字段。
type FbsdkConfirm struct {
	SeqNum       uint64
	TransactTime int64
	ClOrdID      uint64
	OrigClOrdID  uint64
	OrderQty     int64
	LeavesQty    int64
	ExecType     uint8
	OrdStatus    uint8
	Market       uint8
	Side         uint8
	HasRejReason bool
	RejReason    uint16
	SecurityID   uint32
}

// FbsdkTrade —— 成交回报字段。
type FbsdkTrade struct {
	SeqNum         uint64
	TransactTime   int64
	ClOrdID        uint64
	ExecID         uint64
	LastQty        int64
	CumQty         int64
	LeavesQty      int64
	LastPxMantissa int64
	LastPxScale    int8
	SecurityID     uint32
	Market         uint8
	Side           uint8
	OrdStatus      uint8
}

// FbsdkCancelReject —— 撤单失败字段。
type FbsdkCancelReject struct {
	SeqNum       uint64
	TransactTime int64
	ClOrdID      uint64
	OrigClOrdID  uint64
	CxlRejReason uint16
	OrdStatus    uint8
	Market       uint8
	SecurityID   uint32
}

// FbsdkMessage —— fbsdk 事件/期望的统一包装（实现 BinaryCodec 以接入
// 通用执行器；字节编解码由 Rust 核心承担，此处为显式占位）。
type FbsdkMessage struct {
	MsgType      uint32
	NewOrder     *FbsdkNewOrder
	Cancel       *FbsdkCancel
	Confirm      *FbsdkConfirm
	Trade        *FbsdkTrade
	CancelReject *FbsdkCancelReject
}

// Encode 占位：fbsdk 走 SDK 请求通道，不经 Go 线路编码。
func (m *FbsdkMessage) Encode(buf *bytes.Buffer) error {
	return errors.New("fbsdk: 线路编码由 SDK 核心（fbsdk-ffi）实现")
}

// Decode 占位：fbsdk 事件经 SPI 回调入队，不经 Go 线路解码。
func (m *FbsdkMessage) Decode(buf *bytes.Buffer) error {
	return errors.New("fbsdk: 线路解码由 SDK 核心（fbsdk-ffi）实现")
}

// FbsdkMessageCodec —— 用例期望构造 + MsgType 识别。
type FbsdkMessageCodec struct{}

func (c *FbsdkMessageCodec) ProtoName() string { return Fbsdk }

func (c *FbsdkMessageCodec) EncodeJSONMap(map[string]interface{}) ([]byte, error) {
	return nil, errors.New("fbsdk: 线路编码由 SDK 核心（fbsdk-ffi）实现")
}

// JSONToStruct —— 用例期望值构造：按 MsgType 分派到对应字段结构体。
func (c *FbsdkMessageCodec) JSONToStruct(message map[string]interface{}) (fin_codec.BinaryCodec, error) {
	msgType, err := msgTypeFromMap(message)
	if err != nil {
		return nil, err
	}
	msg := &FbsdkMessage{MsgType: msgType}
	switch msgType {
	case FbsdkMsgNewOrder:
		msg.NewOrder = &FbsdkNewOrder{}
		if err := ConvertMapToStruct(message, msg.NewOrder); err != nil {
			return nil, err
		}
	case FbsdkMsgCancel:
		msg.Cancel = &FbsdkCancel{}
		if err := ConvertMapToStruct(message, msg.Cancel); err != nil {
			return nil, err
		}
	case FbsdkMsgOrderConfirm:
		msg.Confirm = &FbsdkConfirm{}
		if err := ConvertMapToStruct(message, msg.Confirm); err != nil {
			return nil, err
		}
	case FbsdkMsgTrade:
		msg.Trade = &FbsdkTrade{}
		if err := ConvertMapToStruct(message, msg.Trade); err != nil {
			return nil, err
		}
	case FbsdkMsgCancelReject:
		msg.CancelReject = &FbsdkCancelReject{}
		if err := ConvertMapToStruct(message, msg.CancelReject); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("fbsdk: 不支持的消息类型 %d", msgType)
	}
	return msg, nil
}

func (c *FbsdkMessageCodec) Encode(uint32, fin_codec.BinaryCodec) ([]byte, error) {
	return nil, errors.New("fbsdk: 线路编码由 SDK 核心（fbsdk-ffi）实现")
}

func (c *FbsdkMessageCodec) Decode([]byte) (interface{}, fin_codec.BinaryCodec, error) {
	return nil, nil, errors.New("fbsdk: 线路解码由 SDK 核心（fbsdk-ffi）实现")
}

// MsgTypeFromMap —— 从用例字段面提取 MsgType（供 fbsdk 模拟器分派请求）。
func MsgTypeFromMap(message map[string]interface{}) (uint32, error) {
	return msgTypeFromMap(message)
}

func init() {
	FbsdkCodecHook = func() (MessageCodec, error) { return &FbsdkMessageCodec{}, nil }
	FbsdkFramerHook = func() (Framer, error) {
		return nil, errors.New("fbsdk: SDK 直连柜台，无 Go 侧帧读取")
	}
}
