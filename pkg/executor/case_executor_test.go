package executor

import (
	"testing"
	"time"

	"github.com/xinchentechnote/gt-auto/pkg/config"
)

// TestReceiveTimeoutFromConfig verifies the receive timeout comes from the
// config's receive_timeout_ms and falls back to the 5s default when unset.
func TestReceiveTimeoutFromConfig(t *testing.T) {
	e := NewCaseExecutor(config.GwAutoConfig{}, nil)
	if e.receiveTimeout != defaultReceiveTimeout {
		t.Fatalf("expected default timeout %v, got %v", defaultReceiveTimeout, e.receiveTimeout)
	}

	e = NewCaseExecutor(config.GwAutoConfig{ReceiveTimeoutMs: 3000}, nil)
	if e.receiveTimeout != 3*time.Second {
		t.Fatalf("expected 3s timeout from config, got %v", e.receiveTimeout)
	}
}
