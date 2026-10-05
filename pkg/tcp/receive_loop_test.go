package tcp

import (
	"bytes"
	"net"
	"testing"
	"time"

	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
)

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
