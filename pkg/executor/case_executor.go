package executor

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/xinchentechnote/gt-auto/pkg/config"
	"github.com/xinchentechnote/gt-auto/pkg/proto"
	"github.com/xinchentechnote/gt-auto/pkg/tcp"
	"github.com/xinchentechnote/gt-auto/pkg/testcase"
	"github.com/xinchentechnote/gt-auto/pkg/validate"
)

// CaseExecutor is responsible for executing test cases.
type CaseExecutor struct {
	Cases        []testcase.TestCase
	Config       config.GwAutoConfig
	simulatorMap map[string]tcp.Simulator[proto.Message]
}

// NewCaseExecutor creates a new CaseExecutor instance.
func NewCaseExecutor(config config.GwAutoConfig, cases []testcase.TestCase) *CaseExecutor {
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
}

func (e *CaseExecutor) executeCase(index int, c testcase.TestCase) {
	log.Infof("Start to execute case: %d, %s - %s\n", index, c.CaseNo, c.CaseTitle)
	for i, step := range c.Steps {
		e.executeStep(i, step)
	}
}

func (e *CaseExecutor) executeStep(index int, step testcase.TestStep) {
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
		actual, err := simulator.Receive()
		if nil != err {
			//TODO
			log.Error("Receive failed: ", err)
			return
		}
		log.Info("Receive data: ", actual)
		step.SetActual(actual)
		result, err := step.Validate()
		if nil != err {
			return
		}
		validate.PrintCompareResult(result)
	}
}
