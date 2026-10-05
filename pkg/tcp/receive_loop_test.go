package tcp

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
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
		queue:  goconcurrentqueue.NewFIFO(),
	}

	start := time.Now()
	_, err := sim.Receive(100 * time.Millisecond)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
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
		queue:  goconcurrentqueue.NewFIFO(),
	}

	if _, err := sim.Receive(100 * time.Millisecond); err == nil {
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
		queue:  goconcurrentqueue.NewFIFO(),
	}

	if _, err := sim.Receive(50 * time.Millisecond); err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	done := make(chan struct{})
	go func() {
		sim.queue.Enqueue(&dummyFrame{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Enqueue blocked after a timed-out Receive")
	}
	msg, err := sim.Receive(time.Second)
	if err != nil {
		t.Fatalf("Receive after timeout failed: %v", err)
	}
	if _, ok := msg.(*dummyFrame); !ok {
		t.Fatalf("unexpected message type %T", msg)
	}
}

// TestOmsReceiveReturnsEnqueuedMessage verifies Receive hands back the queued
// message when one is available.
func TestOmsReceiveReturnsEnqueuedMessage(t *testing.T) {
	riskCodec, framer := newTestSimulatorCodec()
	sim := &OmsSimulator[fin_codec.BinaryCodec]{
		Codec:  riskCodec,
		Framer: framer,
		queue:  goconcurrentqueue.NewFIFO(),
	}
	if err := sim.queue.Enqueue(&dummyFrame{}); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}
	msg, err := sim.Receive(time.Second)
	if err != nil {
		t.Fatalf("Receive failed: %v", err)
	}
	if _, ok := msg.(*dummyFrame); !ok {
		t.Fatalf("unexpected message type %T", msg)
	}
}
