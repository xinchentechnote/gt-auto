package executor

import (
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
)

// CaseExecutor is responsible for executing test cases.
type CaseExecutor struct {
	Cases        []*testcase.TestCase
	Config       config.GwAutoConfig
	simulatorMap map[string]tcp.Simulator[proto.Message]
}

// NewCaseExecutor creates a new CaseExecutor instance.
func NewCaseExecutor(config config.GwAutoConfig, cases []*testcase.TestCase) *CaseExecutor {
	executor := &CaseExecutor{
		Cases:        cases,
		Config:       config,
		simulatorMap: make(map[string]tcp.Simulator[proto.Message]),
	}
	executor.initSimulator()
	return executor
}

func (e *CaseExecutor) initSimulator() {
	for _, config := range e.Config.Simulators {
		simulator, err := tcp.CreateSimulator[proto.Message](config)
		if nil != err {
			continue
		}
		time.Sleep(1000 * time.Millisecond)
		go func() {
			err = simulator.Start()
			if nil != err {
				return
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
	log.Infof("Show to case result: %d, %s - %s\n", index, c.CaseNo, c.CaseTitle)
	for _, result := range c.ValidateResults {
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
	log.Infof("Start to execute case: %d, %s - %s\n", index, c.CaseNo, c.CaseTitle)
	for i, step := range c.Steps {
		e.executeStep(i, c, &step)
	}
}

func (e *CaseExecutor) executeStep(index int, c *testcase.TestCase, step *testcase.TestStep) {
	log.Infof("Start to execute step: %d, %s\n", index, step.StepID)
	var simulator = e.simulatorMap[step.TestTool]
	if nil == simulator {
		conf := e.Config.SimulatorMap[step.TestTool]
		var err error
		simulator, err = tcp.CreateSimulator[proto.Message](conf)
		if nil != err {
			return
		}
		go func() {
			err = simulator.Start()
			if nil != err {
				return
			}
		}()
		e.simulatorMap[step.TestTool] = simulator
	}
	time.Sleep(1000 * time.Millisecond)
	switch step.ActionType {
	case "Send":
		step.TestData["MsgType"] = step.TestFunction
		log.Info("Send data: ", step.TestData)
		err := simulator.SendFromJSON(step.TestData)
		if nil != err {
			log.Errorf("Send failed:%s", err)
		}
	case "Recieve":
		step.TestData["MsgType"] = step.TestFunction
		expect, err := simulator.GetCodec().JSONToStruct(step.TestData)
		if nil != err {
			//TODO
			log.Error("Expect JsonToStruct failed: ", err)
			return
		}
		step.SetExpect(expect)
		actual, err := simulator.Receive()
		if nil != err {
			//TODO
			log.Error("Receive failed: ", err)
			return
		}
		log.Info("TestData data: ", step.TestData)
		log.Info("Actual data: ", actual)
		step.SetActual(actual)
		log.Info("Expected data: ", step.Expect)
		result := step.Validate()

		c.AddValidateResult(index, step.StepID, result)
	}
}
