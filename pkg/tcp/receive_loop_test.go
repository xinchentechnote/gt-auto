package tcp

import (
	"net"
	"testing"
	"time"

	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
)

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
