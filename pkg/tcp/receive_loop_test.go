package tcp

import (
	"bytes"
	"net"
	"testing"
	"time"

	risk_bin "github.com/xinchentechnote/fin-proto-risk-bin-go/messages"
	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
)

func init() {
	// Register a test message type so the risk codec can decode the frames
	// used by the integration test.
	risk_bin.RegistryRcBinaryMsgTypeFactory(999998, func() fin_codec.BinaryCodec {
		return &dummyFrame{}
	})
}

// dummyFrame is a minimal BinaryCodec used to exercise the receive queue.
type dummyFrame struct{}

func (d *dummyFrame) Encode(buf *bytes.Buffer) error { return nil }
func (d *dummyFrame) Decode(buf *bytes.Buffer) error { return nil }

func newTestSimulatorCodec() (*codec.BinaryRiskMessageCodec, *codec.RiskBinFramer) {
	return &codec.BinaryRiskMessageCodec{}, &codec.RiskBinFramer{}
}

// TestOmsReceiveLoopExitsWhenConnCloses verifies the receive goroutine stops
// when the connection is closed instead of busy-looping on read errors.
func TestOmsReceiveLoopExitsWhenConnCloses(t *testing.T) {
	client, server := net.Pipe()
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		conn:   server,
	}

	done := make(chan struct{})
	go func() {
		sim.receiveLoop()
		close(done)
	}()

	client.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("receiveLoop did not exit after connection close")
	}
}

// TestTgwHandleClientExitsWhenClientCloses verifies handleClient returns when
// the client disconnects instead of spinning on read errors.
func TestTgwHandleClientExitsWhenClientCloses(t *testing.T) {
	client, server := net.Pipe()
	riskCodec, framer := newTestSimulatorCodec()
	sim := &TgwSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
	}

	done := make(chan struct{})
	go func() {
		sim.handleClient(server)
		close(done)
	}()

	client.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handleClient did not exit after client disconnect")
	}
}

// TestOmsReceiveTimeout verifies Receive gives up after the timeout instead of
// blocking forever when no message arrives.
func TestOmsReceiveTimeout(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		queue: make(chan receivedMessage, 16),
	}

	start := time.Now()
	if _, _, err := sim.Receive(100 * time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Fatalf("Receive blocked for %v, expected ~100ms", elapsed)
	}
}

// TestTgwReceiveTimeout verifies the TGW simulator honors the timeout too.
func TestTgwReceiveTimeout(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &TgwSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		queue: make(chan receivedMessage, 16),
	}

	if _, _, err := sim.Receive(100 * time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestOmsReceiveAfterTimeout verifies a timed-out Receive does not break
// later enqueues and receives (the queue's watcher cleanup path).
func TestOmsReceiveAfterTimeout(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		queue: make(chan receivedMessage, 16),
	}

	if _, _, err := sim.Receive(50 * time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	done := make(chan struct{})
	go func() {
		sim.queue <- receivedMessage{MsgType: 1, Body: &dummyFrame{}}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Enqueue blocked after a timed-out Receive")
	}
	msg, msgType, err := sim.Receive(time.Second)
	if err != nil {
		t.Fatalf("Receive after timeout failed: %v", err)
	}
	if msgType != 1 {
		t.Fatalf("expected msg type 1, got %d", msgType)
	}
	if _, ok := msg.(*dummyFrame); !ok {
		t.Fatalf("unexpected message type %T", msg)
	}
}

// TestOmsReceiveReturnsEnqueuedMessage verifies Receive hands back the queued
// message and its wire type when one is available.
func TestOmsReceiveReturnsEnqueuedMessage(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		queue: make(chan receivedMessage, 16),
	}
	sim.queue <- receivedMessage{MsgType: 100101, Body: &dummyFrame{}}
	msg, msgType, err := sim.Receive(time.Second)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	if msgType != 100101 {
		t.Fatalf("expected msg type 100101, got %d", msgType)
	}
	if _, ok := msg.(*dummyFrame); !ok {
		t.Fatalf("unexpected message type %T", msg)
	}
}

// TestTgwCloseClosesClientConn verifies Close drops the accepted client
// connection so handleClient can exit.
func TestTgwCloseClosesClientConn(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &TgwSimulator[fin_codec.BinaryCodec]{
		ListenAddress: "127.0.0.1:0",
		Codec:         riskCodec,
		Framer:        framer,
	}
	done := make(chan error, 1)
	go func() { done <- sim.Start() }()

	// Wait for the listener to come up.
	deadline := time.Now().Add(2 * time.Second)
	var addr string
	for time.Now().Before(deadline) {
		sim.stopMu.Lock()
		listener := sim.listener
		sim.stopMu.Unlock()
		if listener != nil {
			addr = listener.Addr().String()
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if addr == "" {
		t.Fatal("TGW listener did not start in time")
	}

	client, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("client dial failed: %v", err)
	}
	defer client.Close()

	// Wait for the server side to register the connection.
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		sim.connMu.Lock()
		registered := sim.conn != nil
		sim.connMu.Unlock()
		if registered {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if err := sim.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	_ = client.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 1)
	if _, err := client.Read(buf); err == nil {
		t.Fatal("expected read error after Close, got none")
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start accept loop did not exit after Close")
	}
}

// TestTgwSendByteWithoutClient verifies sendByte reports an error instead of
// panicking when no client has connected.
func TestTgwSendByteWithoutClient(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &TgwSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
	}
	if err := sim.sendByte([]byte("x")); err == nil {
		t.Fatal("expected error for missing client connection")
	}
}

// TestOmsCloseBeforeStart verifies Close does not panic when Start was never
// called (nil connection).
func TestOmsCloseBeforeStart(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
	}
	if err := sim.Close(); err != nil {
		t.Fatalf("Close before Start should be a no-op, got: %v", err)
	}
}

// TestOmsTgwEndToEnd wires an OMS client to a TGW server over the risk
// protocol and verifies both directions: OMS.Send -> TGW.Receive and
// TGW.SendFromJSON -> OMS.Receive.
func TestOmsTgwEndToEnd(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	tgw := &TgwSimulator[fin_codec.BinaryCodec]{
		ListenAddress: "127.0.0.1:0",
		queue:         make(chan receivedMessage, 16),
		Codec:         riskCodec,
		Framer:        framer,
	}
	go func() { _ = tgw.Start() }()
	t.Cleanup(func() { _ = tgw.Close() })
	waitForReady(t, tgw)

	oms := &OmsSimulator[fin_codec.BinaryCodec]{
		ServerAddress: tgw.listener.Addr().String(),
		queue:         make(chan receivedMessage, 16),
		Codec:         riskCodec,
		Framer:        framer,
	}
	if err := oms.Start(); err != nil {
		t.Fatalf("OMS start failed: %v", err)
	}
	t.Cleanup(func() { _ = oms.Close() })
	waitForReady(t, oms)

	// OMS -> TGW
	if err := oms.Send(999998, &dummyFrame{}); err != nil {
		t.Fatalf("OMS send failed: %v", err)
	}
	body, msgType, err := tgw.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("TGW receive failed: %v", err)
	}
	if msgType != 999998 {
		t.Fatalf("TGW got msg type %d, want 999998", msgType)
	}
	if _, ok := body.(*dummyFrame); !ok {
		t.Fatalf("TGW got unexpected body type %T", body)
	}

	// TGW -> OMS through the JSON encoding path
	if err := tgw.SendFromJSON(map[string]interface{}{"MsgType": "999998"}); err != nil {
		t.Fatalf("TGW send failed: %v", err)
	}
	body, msgType, err = oms.Receive(2 * time.Second)
	if err != nil {
		t.Fatalf("OMS receive failed: %v", err)
	}
	if msgType != 999998 {
		t.Fatalf("OMS got msg type %d, want 999998", msgType)
	}
	if _, ok := body.(*dummyFrame); !ok {
		t.Fatalf("OMS got unexpected body type %T", body)
	}
}

// waitForReady blocks until the simulator reports ready or the test times out.
func waitForReady(t *testing.T, sim Simulator[fin_codec.BinaryCodec]) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sim.Ready() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("simulator did not become ready in time")
}
