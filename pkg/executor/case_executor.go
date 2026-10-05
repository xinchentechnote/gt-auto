package executor

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/xinchentechnote/fin-proto-runtime-bin-go/codec"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
)

// defaultReceiveTimeout is used when the config does not set
// receive_timeout_ms.
const defaultReceiveTimeout = 5 * time.Second

// CaseExecutor is responsible for executing test cases.
type CaseExecutor struct {
	Cases          []*testcase.TestCase
	Config         config.GwAutoConfig
	simulatorMap   map[string]tcp.Simulator[codec.BinaryCodec]
	receiveTimeout time.Duration
}

// NewCaseExecutor creates a new CaseExecutor instance.
func NewCaseExecutor(config config.GwAutoConfig, cases []*testcase.TestCase) *CaseExecutor {
	receiveTimeout := defaultReceiveTimeout
	if config.ReceiveTimeoutMs > 0 {
		receiveTimeout = time.Duration(config.ReceiveTimeoutMs) * time.Millisecond
	}
	executor := &CaseExecutor{
		Cases:          cases,
		Config:         config,
		simulatorMap:   make(map[string]tcp.Simulator[codec.BinaryCodec]),
		receiveTimeout: receiveTimeout,
	}
	executor.initSimulator()
	return executor
}

func (e *CaseExecutor) initSimulator() {
	for _, config := range e.Config.Simulators {
		simulator, err := tcp.CreateSimulator[codec.BinaryCodec](config)
		if nil != err {
			log.Errorf("Failed to create simulator %s: %v", config.Name, err)
			continue
		}
		time.Sleep(1000 * time.Millisecond)
		go func() {
			if startErr := simulator.Start(); startErr != nil {
				log.Errorf("Failed to start simulator %s: %v", config.Name, startErr)
			}
		}()
		e.simulatorMap[config.Name] = simulator
	}
	time.Sleep(1000 * time.Millisecond)
}

// Execute runs the test cases.
func (e *CaseExecutor) Execute() {
	time.Sleep(5 * time.Second)
	if e.Cases == nil {
		return
	}
	for i, c := range e.Cases {
		e.executeCase(i, c)
	}

	for i, c := range e.Cases {
		e.showResult(i, c)
	}
}

func (e *CaseExecutor) showResult(index int, c *testcase.TestCase) {
	log.Infof("Show to case result: %d, %s - %s\n", index, c.CaseID, c.CaseTitle)
	for _, result := range c.ValidateResults {
		if result.Error != "" {
			log.Errorf("Show to case result: %d, %s❌ step failed: %s", result.Index, result.StepID, result.Error)
			continue
		}
		if !result.Passed {
			log.Errorf("Show to case result: %d, %s❌", result.Index, result.StepID)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Path", "Expected", "Actual"})
			for _, diff := range result.Detail.Diffs {
				table.Append([]string{
					diff.Path,
					fmt.Sprintf("%v", diff.Expect),
					fmt.Sprintf("%v", diff.Actual),
				})
			}
			table.Render()
		} else {
			log.Infof("Show to case result: %d-%s:✅", result.Index, result.StepID)
		}
	}
}

func (e *CaseExecutor) executeCase(index int, c *testcase.TestCase) {
	log.Infof("Start to execute case: %d, %s - %s\n", index, c.CaseID, c.CaseTitle)
	for i, step := range c.Steps {
		e.executeStep(i, c, &step)
	}
}

func (e *CaseExecutor) executeStep(index int, c *testcase.TestCase, step *testcase.TestStep) {
	log.Infof("Start to execute step: %d, %s\n", index, step.StepID)
	simulator, err := e.getOrStartSimulator(step.TestTool)
	if err != nil {
		log.Errorf("Step %d-%s cannot run: %v", index, step.StepID, err)
		c.AddStepError(index, step.StepID, err)
		return
	}
	time.Sleep(1000 * time.Millisecond)
	switch step.ActionType {
	case "Send":
		step.TestDatas["MsgType"] = step.MsgType
		log.Info("Send data: ", step.TestDatas)
		if err := simulator.SendFromJSON(step.TestDatas); err != nil {
			log.Errorf("Send failed:%s", err)
			c.AddStepError(index, step.StepID, fmt.Errorf("send failed: %w", err))
		}
	case "Receive":
		step.TestDatas["MsgType"] = step.MsgType
		expect, err := simulator.GetCodec().JSONToStruct(step.TestDatas)
		if err != nil {
			//TODO
			log.Error("Expect JsonToStruct failed: ", err)
			c.AddStepError(index, step.StepID, fmt.Errorf("build expected message: %w", err))
			return
		}
		step.SetExpect(expect)
		actual, err := simulator.Receive(e.receiveTimeout)
		if err != nil {
			//TODO
			log.Error("Receive failed: ", err)
			c.AddStepError(index, step.StepID, fmt.Errorf("receive failed: %w", err))
			return
		}
		if step.VerifyRequired {
			log.Info("TestData data: ", step.TestDatas)
			log.Info("Actual data: ", actual)
			step.SetActual(actual)
			log.Info("Expected data: ", step.Expect)
			result := step.Validate()
			c.AddValidateResult(index, step.StepID, result)
		}
	default:
		err := fmt.Errorf("unknown action type: %s", step.ActionType)
		log.Warn(err)
		c.AddStepError(index, step.StepID, err)
	}
}

// getOrStartSimulator returns the simulator for the test tool, lazily
// creating and starting it on first use.
func (e *CaseExecutor) getOrStartSimulator(testTool string) (tcp.Simulator[codec.BinaryCodec], error) {
	if simulator, ok := e.simulatorMap[testTool]; ok {
		return simulator, nil
	}
	conf, ok := e.Config.SimulatorMap[testTool]
	if !ok {
		return nil, fmt.Errorf("test tool %q not found in config simulators", testTool)
	}
	simulator, err := tcp.CreateSimulator[codec.BinaryCodec](conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create simulator %q: %w", testTool, err)
	}
	go func() {
		if startErr := simulator.Start(); startErr != nil {
			log.Errorf("Failed to start simulator %s: %v", testTool, startErr)
		}
	}()
	e.simulatorMap[testTool] = simulator
	return simulator, nil
}
