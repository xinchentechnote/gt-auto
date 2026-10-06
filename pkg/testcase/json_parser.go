package testcase

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// JSONCaseParser implements the CaseParser interface for JSON files.
//
// Unlike the CSV format, the JSON format carries each step's test data
// inline ("testData" object), so no separate data sheets are needed.
// See docs/design.md for the format spec.
type JSONCaseParser struct {
	FilePath string
}

type jsonTestDataFile struct {
	Cases []jsonTestCase `json:"cases"`
}

type jsonTestCase struct {
	CaseID    string     `json:"caseId"`
	CaseTitle string     `json:"caseTitle"`
	Steps     []jsonStep `json:"steps"`
}

type jsonStep struct {
	StepID         string `json:"stepId"`
	StepDesc       string `json:"stepDesc"`
	ActionType     string `json:"actionType"`
	VerifyRequired bool   `json:"verifyRequired"`
	TestTool       string `json:"testTool"`
	// SleepMs and MsgType accept both strings and numbers for convenience
	// and are normalized to the model's string representation.
	SleepMs   any            `json:"sleepMs"`
	MsgType   any            `json:"msgType"`
	TestDatas map[string]any `json:"testData"`
}

// Parse parses a JSON case file and returns the test cases.
func (p *JSONCaseParser) Parse() ([]*TestCase, error) {
	data, err := os.ReadFile(p.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read case file: %w", err)
	}
	var file jsonTestDataFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("failed to parse JSON case file %s: %w", p.FilePath, err)
	}

	var cases []*TestCase
	for _, jc := range file.Cases {
		tc := &TestCase{
			CaseID:    jc.CaseID,
			CaseTitle: jc.CaseTitle,
			Steps:     []TestStep{},
		}
		for _, js := range jc.Steps {
			sleepMs, err := anyToString(js.SleepMs)
			if err != nil {
				return nil, fmt.Errorf("case %s step %s: sleepMs: %w", jc.CaseID, js.StepID, err)
			}
			msgType, err := anyToString(js.MsgType)
			if err != nil {
				return nil, fmt.Errorf("case %s step %s: msgType: %w", jc.CaseID, js.StepID, err)
			}
			testDatas := js.TestDatas
			if testDatas == nil {
				testDatas = map[string]any{}
			}
			tc.Steps = append(tc.Steps, TestStep{
				StepID:         js.StepID,
				SleepMs:        sleepMs,
				StepDesc:       js.StepDesc,
				ActionType:     js.ActionType,
				VerifyRequired: js.VerifyRequired,
				TestTool:       js.TestTool,
				MsgType:        msgType,
				TestDatas:      testDatas,
			})
		}
		cases = append(cases, tc)
	}
	return cases, nil
}

// anyToString normalizes a decoded JSON scalar to its string representation.
func anyToString(v any) (string, error) {
	switch t := v.(type) {
	case nil:
		return "", nil
	case string:
		return t, nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("must be a string or a number, got %T", v)
	}
}
