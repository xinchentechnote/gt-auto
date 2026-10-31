//go:build fbsdk

// fbsdk 客户端模拟器（-tags fbsdk 构建）：以真实 FB-SDK（Go 绑定 → Rust 核心）
// 作为被测柜台的客户端侧，经 cgo 驱动 fbsdk-ffi。
//
// 与通用 OMS/TGW 模拟器的差异：会话（登录/心跳/重连/续传）与线格式编解码全部
// 由 SDK 核心承担——Send 走 SDK 请求接口，Receive 来自 SPI 回调入队。
package tcp

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xinchentechnote/fbsdk/go/fbsdk"
	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	gtcodec "github.com/xinchentechnote/gt-auto/pkg/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
)

// FbsdkSimulator —— FB-SDK 客户端模拟器（fcounter ingress 侧）。
type FbsdkSimulator struct {
	cfg      config.SimulatorConfig
	api      *fbsdk.TraderApi
	queue    chan *gtcodec.FbsdkMessage
	ready    atomic.Bool
	closed   atomic.Bool
	started  atomic.Bool
	dropped  atomic.Int64
	startErr error
	mu       sync.Mutex
}

// spiBridge —— fbsdk.Spi 回调（Rust 回调线程）→ 事件队列。
type spiBridge struct {
	sim *FbsdkSimulator
}

func (b *spiBridge) OnFrontConnected() { b.sim.ready.Store(true) }
func (b *spiBridge) OnDisconnected(reason string) {
	b.sim.ready.Store(false)
}

func (b *spiBridge) OnOrderConfirm(c *fbsdk.OrderConfirm) {
	if !b.sim.enqueue(&gtcodec.FbsdkMessage{
		MsgType: gtcodec.FbsdkMsgOrderConfirm,
		Confirm: &gtcodec.FbsdkConfirm{
			SeqNum:       c.SeqNum,
			TransactTime: c.TransactTime,
			ClOrdID:      c.ClOrdID,
			OrigClOrdID:  c.OrigClOrdID,
			OrderQty:     c.OrderQty,
			LeavesQty:    c.LeavesQty,
			ExecType:     c.ExecType,
			OrdStatus:    c.OrdStatus,
			Market:       c.Market,
			Side:         c.Side,
			HasRejReason: c.HasRejReason,
			RejReason:    c.RejReason,
			SecurityID:   c.SecurityID,
		},
	}) {
		b.sim.dropped.Add(1)
	}
}

func (b *spiBridge) OnTrade(t *fbsdk.Trade) {
	if !b.sim.enqueue(&gtcodec.FbsdkMessage{
		MsgType: gtcodec.FbsdkMsgTrade,
		Trade: &gtcodec.FbsdkTrade{
			SeqNum:         t.SeqNum,
			TransactTime:   t.TransactTime,
			ClOrdID:        t.ClOrdID,
			ExecID:         t.ExecID,
			LastQty:        t.LastQty,
			CumQty:         t.CumQty,
			LeavesQty:      t.LeavesQty,
			LastPxMantissa: t.LastPx.Mantissa,
			LastPxScale:    t.LastPx.Scale,
			SecurityID:     t.SecurityID,
			Market:         t.Market,
			Side:           t.Side,
			OrdStatus:      t.OrdStatus,
		},
	}) {
		b.sim.dropped.Add(1)
	}
}

func (b *spiBridge) OnCancelReject(r *fbsdk.CancelReject) {
	if !b.sim.enqueue(&gtcodec.FbsdkMessage{
		MsgType: gtcodec.FbsdkMsgCancelReject,
		CancelReject: &gtcodec.FbsdkCancelReject{
			SeqNum:       r.SeqNum,
			TransactTime: r.TransactTime,
			ClOrdID:      r.ClOrdID,
			OrigClOrdID:  r.OrigClOrdID,
			CxlRejReason: r.CxlRejReason,
			OrdStatus:    r.OrdStatus,
			Market:       r.Market,
			SecurityID:   r.SecurityID,
		},
	}) {
		b.sim.dropped.Add(1)
	}
}

func (b *spiBridge) OnQueryAccountRsp(r *fbsdk.AccountRsp) {}
func (b *spiBridge) OnQueryPositionRsp(r *fbsdk.PositionRsp) {
}
func (b *spiBridge) OnQueryOrderRsp(r *fbsdk.OrderQryRsp) {}
func (b *spiBridge) OnQueryTradeRsp(r *fbsdk.TradeQryRsp) {}

func (s *FbsdkSimulator) enqueue(m *gtcodec.FbsdkMessage) bool {
	select {
	case s.queue <- m:
		return true
	default:
		return false
	}
}

// Start —— 创建 SDK 实例并发起连接（异步；Ready() 轮询连接就绪）。
func (s *FbsdkSimulator) Start() error {
	if s.started.Swap(true) {
		return fmt.Errorf("fbsdk 模拟器重复 Start")
	}
	api := fbsdk.Create()
	api.RegisterSpi(&spiBridge{sim: s})
	clientID := s.cfg.ClientID
	if clientID == "" {
		clientID = "gt-auto"
	}
	cfg := fbsdk.Config{ClientID: clientID, Token: s.cfg.Token}
	if err := api.Init(s.cfg.ServerAddress, cfg); err != nil {
		s.startErr = err
		return fmt.Errorf("fbsdk Init(%s): %w", s.cfg.ServerAddress, err)
	}
	s.mu.Lock()
	s.api = api
	s.mu.Unlock()
	return nil
}

// Ready —— 连接就绪（OnFrontConnected 已触发）。
func (s *FbsdkSimulator) Ready() bool { return s.ready.Load() }

// Send —— 按 MsgType 分派 SDK 请求接口。
func (s *FbsdkSimulator) Send(msgType uint32, _ fin_codec.BinaryCodec) error {
	return errors.New("fbsdk: 请使用 SendFromJSON（字段面构造请求）")
}

// SendFromJSON —— 用例字段面 → SDK 请求。
func (s *FbsdkSimulator) SendFromJSON(message map[string]interface{}) error {
	if s.api == nil {
		return errors.New("fbsdk: SDK 未启动")
	}
	msgType, err := gtcodec.MsgTypeFromMap(message)
	if err != nil {
		return err
	}
	switch msgType {
	case gtcodec.FbsdkMsgNewOrder:
		var o gtcodec.FbsdkNewOrder
		if err := gtcodec.ConvertMapToStruct(message, &o); err != nil {
			return err
		}
		return s.api.ReqOrderInsert(&fbsdk.NewOrderSingle{
			ClOrdID:    o.ClOrdID,
			OrderQty:   o.OrderQty,
			Price:      fbsdk.NewDecimal(o.PriceMantissa, o.PriceScale),
			SecurityID: o.SecurityID,
			AccountID:  o.AccountID,
			Market:     o.Market,
			Side:       o.Side,
			OrdType:    o.OrdType,
		})
	case gtcodec.FbsdkMsgCancel:
		var c gtcodec.FbsdkCancel
		if err := gtcodec.ConvertMapToStruct(message, &c); err != nil {
			return err
		}
		return s.api.ReqOrderCancel(&fbsdk.OrderCancel{
			ClOrdID:     c.ClOrdID,
			OrigClOrdID: c.OrigClOrdID,
			SecurityID:  c.SecurityID,
			AccountID:   c.AccountID,
			Market:      c.Market,
			Side:        c.Side,
		})
	case gtcodec.FbsdkMsgQueryAccount:
		var q struct{ AccountID uint32 }
		if err := gtcodec.ConvertMapToStruct(message, &q); err != nil {
			return err
		}
		return s.api.ReqQueryAccount(q.AccountID)
	case gtcodec.FbsdkMsgQueryPosition:
		var q struct {
			AccountID  uint32
			SecurityID uint32
		}
		if err := gtcodec.ConvertMapToStruct(message, &q); err != nil {
			return err
		}
		return s.api.ReqQueryPosition(q.AccountID, q.SecurityID)
	case gtcodec.FbsdkMsgQueryOrder:
		var q struct{ ClOrdID uint64 }
		if err := gtcodec.ConvertMapToStruct(message, &q); err != nil {
			return err
		}
		return s.api.ReqQueryOrder(q.ClOrdID)
	case gtcodec.FbsdkMsgQueryTrade:
		var q struct {
			ClOrdID uint64
			ExecID  uint64
		}
		if err := gtcodec.ConvertMapToStruct(message, &q); err != nil {
			return err
		}
		return s.api.ReqQueryTrade(q.ClOrdID, q.ExecID)
	default:
		return fmt.Errorf("fbsdk: 不支持的消息类型 %d", msgType)
	}
}

// Receive —— 出队一条 SDK 事件。
func (s *FbsdkSimulator) Receive(timeout time.Duration) (fin_codec.BinaryCodec, uint32, error) {
	select {
	case ev := <-s.queue:
		return ev, ev.MsgType, nil
	case <-time.After(timeout):
		return nil, 0, fmt.Errorf("fbsdk: 接收超时（%s）", timeout)
	}
}

// GetCodec —— 期望值构造器（字段面）。
func (s *FbsdkSimulator) GetCodec() gtcodec.MessageCodec {
	return &gtcodec.FbsdkMessageCodec{}
}

// Close —— 停止 SDK 并释放（幂等）。
func (s *FbsdkSimulator) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	s.mu.Lock()
	api := s.api
	s.api = nil
	s.mu.Unlock()
	if api != nil {
		api.Close()
	}
	return nil
}

// 注册 fbsdk 模拟器钩子（-tags fbsdk 构建时生效；默认构建 factory 返回指引性错误）。
func init() {
	FbsdkSimulatorHook = func(config config.SimulatorConfig) (any, error) {
		return &FbsdkSimulator{
			cfg:   config,
			queue: make(chan *gtcodec.FbsdkMessage, 1024),
		}, nil
	}
}
