//go:build fbsdk

package codec

import (
	"testing"
)

// fbsdk 消息面自测（-tags fbsdk）：JSONToStruct 按消息类型分派 + 字段面宽松转换
//（字符串 → 数值/负数/布尔）+ 非法输入快速失败。线格式编解码由 Rust 核心承担，
// 字节级兼容性由 fbsdk 仓的黄金帧对拍测试锁定。

func fbsdkMap(msgType string, kv map[string]string) map[string]interface{} {
	m := map[string]interface{}{"MsgType": msgType}
	for k, v := range kv {
		m[k] = v
	}
	return m
}

// TestFbsdkJSONToStructDispatch —— 各消息类型分派到对应字段结构体并完成宽松转换。
func TestFbsdkJSONToStructDispatch(t *testing.T) {
	c := &FbsdkMessageCodec{}

	// 新委托（64）：负数定点 scale、u64 订单号
	msg, err := c.JSONToStruct(fbsdkMap("64", map[string]string{
		"ClOrdID": "500000001", "SecurityID": "600519", "AccountID": "1",
		"PriceMantissa": "17000000", "PriceScale": "-4", "OrderQty": "100",
		"Market": "1", "Side": "1", "OrdType": "1",
	}))
	if err != nil {
		t.Fatalf("NOS: %v", err)
	}
	nos := msg.(*FbsdkMessage)
	if nos.NewOrder == nil || nos.NewOrder.ClOrdID != 500000001 ||
		nos.NewOrder.PriceMantissa != 17000000 || nos.NewOrder.PriceScale != -4 ||
		nos.NewOrder.OrderQty != 100 || nos.NewOrder.SecurityID != 600519 {
		t.Fatalf("NOS 字段不符: %+v", nos.NewOrder)
	}

	// 委托确认（66）：布尔字段宽松转换
	msg, err = c.JSONToStruct(fbsdkMap("66", map[string]string{
		"SeqNum": "1", "ClOrdID": "500000001", "OrderQty": "100", "LeavesQty": "100",
		"ExecType": "0", "OrdStatus": "0", "HasRejReason": "false",
	}))
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	cf := msg.(*FbsdkMessage)
	if cf.Confirm == nil || cf.Confirm.SeqNum != 1 || cf.Confirm.ExecType != 0 ||
		cf.Confirm.HasRejReason {
		t.Fatalf("Confirm 字段不符: %+v", cf.Confirm)
	}

	// 成交回报（68）
	msg, err = c.JSONToStruct(fbsdkMap("68", map[string]string{
		"ClOrdID": "500000001", "ExecID": "1", "LastQty": "100", "CumQty": "100",
		"LastPxMantissa": "17000000", "LastPxScale": "-4", "OrdStatus": "2",
	}))
	if err != nil {
		t.Fatalf("Trade: %v", err)
	}
	tr := msg.(*FbsdkMessage)
	if tr.Trade == nil || tr.Trade.ExecID != 1 || tr.Trade.LastPxMantissa != 17000000 {
		t.Fatalf("Trade 字段不符: %+v", tr.Trade)
	}

	// 撤单（65）与撤单失败（67）
	if msg, err = c.JSONToStruct(fbsdkMap("65", map[string]string{
		"ClOrdID": "2", "OrigClOrdID": "1",
	})); err != nil || msg.(*FbsdkMessage).Cancel == nil {
		t.Fatalf("Cancel: %v %+v", err, msg)
	}
	if msg, err = c.JSONToStruct(fbsdkMap("67", map[string]string{
		"CxlRejReason": "32", "OrigClOrdID": "999999999",
	})); err != nil || msg.(*FbsdkMessage).CancelReject == nil {
		t.Fatalf("CancelReject: %v %+v", err, msg)
	}
}

// TestFbsdkJSONToStructErrors —— 缺失/非法 MsgType 必须报错而非 panic。
func TestFbsdkJSONToStructErrors(t *testing.T) {
	c := &FbsdkMessageCodec{}
	if _, err := c.JSONToStruct(map[string]interface{}{"ClOrdID": "1"}); err == nil {
		t.Fatal("缺失 MsgType 应报错")
	}
	if _, err := c.JSONToStruct(fbsdkMap("999", nil)); err == nil {
		t.Fatal("非法 MsgType 应报错")
	}
	if _, err := c.JSONToStruct(fbsdkMap("not-a-number", nil)); err == nil {
		t.Fatal("非数值 MsgType 应报错")
	}
}
