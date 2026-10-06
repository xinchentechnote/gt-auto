package codec

import (
	"encoding/binary"
	"net"
	"strings"
	"testing"
)

// TestFramersReadFrameTruncatedBody verifies that a truncated frame body
// surfaces the underlying read error instead of wrapping a nil error.
func TestFramersReadFrameTruncatedBody(t *testing.T) {
	tests := []struct {
		name     string
		framer   Framer
		headSize int
		lenAt    int
	}{
		{"szse", &SzseBinFramer{}, 8, 4},
		{"risk", &RiskBinFramer{}, 12, 8},
		{"sse", &SseBinFramer{}, 16, 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := make([]byte, tt.headSize)
			binary.BigEndian.PutUint32(head[tt.lenAt:tt.lenAt+4], 100)

			server, client := net.Pipe()
			go func() {
				_, _ = client.Write(head)
				_ = client.Close()
			}()

			_, err := tt.framer.ReadFrame(server)
			if err == nil {
				t.Fatal("expected error for truncated body, got nil")
			}
			if strings.Contains(err.Error(), "%!w") {
				t.Fatalf("error lost the wrapped cause: %v", err)
			}
			if !strings.Contains(err.Error(), "EOF") {
				t.Fatalf("expected EOF cause in error, got: %v", err)
			}
		})
	}
}
