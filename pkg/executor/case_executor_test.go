package executor

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	risk_bin "github.com/xinchentechnote/fin-proto-risk-bin-go/messages"
	fin_codec "github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/codec"
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

// TestInitSimulatorStartsAllSimulators verifies all auto-start simulators are
// created and started. Run with -race: the start goroutines used to write the
// loop's err variable, racing with the next CreateSimulator call.
func TestInitSimulatorStartsAllSimulators(t *testing.T) {
	conf := config.GwAutoConfig{
		Simulators: []config.SimulatorConfig{
			{Name: "tgw1", Type: "tgw", Protocol: "binary-risk", ListenAddress: "127.0.0.1:0", AutoStart: true},
			{Name: "tgw2", Type: "tgw", Protocol: "binary-risk", ListenAddress: "127.0.0.1:0", AutoStart: true},
		},
	}
	conf.InitConfigMap()
	e := NewCaseExecutor(conf, nil)
	if len(e.simulatorMap) != 2 {
		t.Fatalf("expected 2 simulators, got %d", len(e.simulatorMap))
	}
	for name, sim := range e.simulatorMap {
		if !sim.Ready() {
			t.Fatalf("simulator %s should be ready after init", name)
		}
		if err := sim.Close(); err != nil {
			t.Fatalf("failed to close simulator %s: %v", name, err)
		}
	}
}

// TestInitSimulatorHonorsAutoStart verifies only auto_start simulators are
// started up front; the rest start lazily on first use.
func TestInitSimulatorHonorsAutoStart(t *testing.T) {
	// A real listener so the lazily started OMS simulator can dial successfully.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind dummy listener: %v", err)
	}
	defer lis.Close()

	conf := config.GwAutoConfig{
		Simulators: []config.SimulatorConfig{
			{Name: "tgw_auto", Type: "tgw", Protocol: "binary-risk", ListenAddress: "127.0.0.1:0", AutoStart: true},
			{Name: "oms_lazy", Type: "oms", Protocol: "binary-risk", ServerAddress: lis.Addr().String(), AutoStart: false},
		},
	}
	conf.InitConfigMap()
	e := NewCaseExecutor(conf, nil)
	defer e.simulatorMap["tgw_auto"].Close()

	if _, ok := e.simulatorMap["tgw_auto"]; !ok {
		t.Fatal("auto_start simulator should be started during init")
	}
	if _, ok := e.simulatorMap["oms_lazy"]; ok {
		t.Fatal("non-auto simulator should not be started during init")
	}

	sim, err := e.getOrStartSimulator("oms_lazy")
	if err != nil {
		t.Fatalf("lazy start failed: %v", err)
	}
	if sim == nil {
		t.Fatal("lazy start returned nil simulator")
	}
	if err := sim.Close(); err != nil {
		t.Fatalf("lazy simulator close failed: %v", err)
	}
}

// TestSleepBeforeStepHonorsSleepMs verifies the step sleeps for its sleep_ms
// value and tolerates empty or invalid values without panicking.
func TestSleepBeforeStepHonorsSleepMs(t *testing.T) {
	step := &testcase.TestStep{SleepMs: "50"}
	start := time.Now()
	sleepBeforeStep(step)
	if elapsed := time.Since(start); elapsed < 50*time.Millisecond {
		t.Fatalf("expected at least 50ms sleep, slept %v", elapsed)
	}

	for _, value := range []string{"", "abc", "0", "-5"} {
		step := &testcase.TestStep{SleepMs: value}
		start := time.Now()
		sleepBeforeStep(step)
		if elapsed := time.Since(start); elapsed > 20*time.Millisecond {
			t.Fatalf("sleep_ms %q should not sleep, took %v", value, elapsed)
		}
	}
}

// TestSummarizeCountsResults verifies the run summary aggregates recorded
// step results across cases.
func TestSummarizeCountsResults(t *testing.T) {
	e := &CaseExecutor{receiveTimeout: time.Second}
	e.Cases = []*testcase.TestCase{
		{
			CaseID: "c1",
			ValidateResults: []testcase.StepValidateResult{
				{StepID: "s1", Passed: true},
				{StepID: "s2", Passed: false, Error: "boom"},
			},
		},
		{
			CaseID: "c2",
			ValidateResults: []testcase.StepValidateResult{
				{StepID: "s3", Passed: true},
			},
		},
	}
	summary := e.summarize()
	if summary.TotalCases != 2 || summary.TotalSteps != 3 ||
		summary.PassedSteps != 2 || summary.FailedSteps != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

// stubSimulator is a controllable Simulator used to exercise the executor's
// Receive-step logic without a network.
type stubSimulator struct {
	msgType uint32
	body    fin_codec.BinaryCodec
	err     error
}

func (s *stubSimulator) Start() error                                             { return nil }
func (s *stubSimulator) Ready() bool                                              { return true }
func (s *stubSimulator) Send(msgType uint32, message fin_codec.BinaryCodec) error { return nil }
func (s *stubSimulator) SendFromJSON(map[string]interface{}) error                { return nil }
func (s *stubSimulator) Close() error                                             { return nil }
func (s *stubSimulator) GetCodec() codec.MessageCodec                             { return &codec.BinaryRiskMessageCodec{} }
func (s *stubSimulator) Receive(timeout time.Duration) (fin_codec.BinaryCodec, uint32, error) {
	return s.body, s.msgType, s.err
}

// TestExecuteStepValidatesReceivedMsgType verifies a Receive step fails with
// a clear error when the wire message type differs from the expected one.
func TestExecuteStepValidatesReceivedMsgType(t *testing.T) {
	tests := []struct {
		name       string
		sim        *stubSimulator
		wantErr    string
		wantPassed bool
	}{
		{
			name:       "matching type validates",
			sim:        &stubSimulator{msgType: 200102, body: &risk_bin.OrderConfirm{}},
			wantPassed: true,
		},
		{
			name:    "mismatched type fails",
			sim:     &stubSimulator{msgType: 999999, body: &risk_bin.OrderConfirm{}},
			wantErr: "received MsgType 999999, expected 200102",
		},
		{
			name:    "receive timeout fails",
			sim:     &stubSimulator{err: fmt.Errorf("no message received within 1s")},
			wantErr: "receive failed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &CaseExecutor{receiveTimeout: time.Second}
			c := &testcase.TestCase{
				CaseID: "c1",
				Steps: []testcase.TestStep{{
					StepID:         "s1",
					TestTool:       "tool",
					ActionType:     "Receive",
					MsgType:        "200102",
					VerifyRequired: true,
					TestDatas:      map[string]any{"StepId": "s1"},
				}},
			}
			e.simulatorMap = map[string]tcp.Simulator[fin_codec.BinaryCodec]{"tool": tt.sim}
			e.executeStep(0, c, &c.Steps[0])

			if len(c.ValidateResults) != 1 {
				t.Fatalf("expected 1 validate result, got %d", len(c.ValidateResults))
			}
			result := c.ValidateResults[0]
			if tt.wantPassed {
				if !result.Passed {
					t.Fatalf("expected step to pass, got error: %s", result.Error)
				}
				return
			}
			if result.Passed {
				t.Fatal("expected step to fail")
			}
			if !strings.Contains(result.Error, tt.wantErr) {
				t.Fatalf("expected error containing %q, got: %s", tt.wantErr, result.Error)
			}
		})
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
