package executor

import (
	"strings"
	"testing"
	"time"

	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
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

// TestExecuteStepRecordsUnavailableSimulator verifies a step whose test tool
// cannot be created is recorded as a failed result instead of being skipped.
func TestExecuteStepRecordsUnavailableSimulator(t *testing.T) {
	conf := config.GwAutoConfig{
		SimulatorMap: map[string]config.SimulatorConfig{
			"bad_tool": {Name: "bad_tool", Type: "no-such-type"},
		},
	}
	e := &CaseExecutor{
		Config:         conf,
		simulatorMap:   make(map[string]tcp.Simulator[fin_codec.BinaryCodec]),
		receiveTimeout: time.Second,
	}
	c := &testcase.TestCase{
		CaseID: "c1",
		Steps:  []testcase.TestStep{{StepID: "s1", TestTool: "bad_tool", ActionType: "Send"}},
	}
	e.executeStep(0, c, &c.Steps[0])

	if len(c.ValidateResults) != 1 {
		t.Fatalf("expected 1 validate result, got %d", len(c.ValidateResults))
	}
	result := c.ValidateResults[0]
	if result.Passed {
		t.Fatal("expected step to be marked as failed")
	}
	if !strings.Contains(result.Error, "bad_tool") {
		t.Fatalf("expected error to mention the test tool, got: %s", result.Error)
	}
}

// TestExecuteStepRecordsUnknownActionType verifies an unrecognized action type
// is recorded as a failed result instead of only logging a warning.
func TestExecuteStepRecordsUnknownActionType(t *testing.T) {
	conf := config.GwAutoConfig{
		SimulatorMap: map[string]config.SimulatorConfig{
			"tool": {Name: "tool", Type: "tgw", Protocol: "binary-risk", ListenAddress: "127.0.0.1:0"},
		},
	}
	sim, err := tcp.CreateSimulator[fin_codec.BinaryCodec](conf.SimulatorMap["tool"])
	if err != nil {
		t.Fatalf("CreateSimulator failed: %v", err)
	}
	e := &CaseExecutor{
		Config:         conf,
		simulatorMap:   map[string]tcp.Simulator[fin_codec.BinaryCodec]{"tool": sim},
		receiveTimeout: time.Second,
	}
	c := &testcase.TestCase{
		CaseID: "c1",
		Steps:  []testcase.TestStep{{StepID: "s1", TestTool: "tool", ActionType: "Bogus"}},
	}
	e.executeStep(0, c, &c.Steps[0])

	if len(c.ValidateResults) != 1 {
		t.Fatalf("expected 1 validate result, got %d", len(c.ValidateResults))
	}
	result := c.ValidateResults[0]
	if result.Passed {
		t.Fatal("expected step to be marked as failed")
	}
	if !strings.Contains(result.Error, "unknown action type") {
		t.Fatalf("expected unknown-action error, got: %s", result.Error)
	}
}
