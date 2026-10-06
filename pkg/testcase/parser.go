package testcase

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CSVCaseParser implements the CaseParser interface for CSV files.
type CSVCaseParser struct {
	FilePath string
	// sheetCache maps a test data sheet name to its records keyed by StepId,
	// so each sheet file is read at most once and identical StepIds in
	// different sheets cannot collide.
	sheetCache map[string]map[string]map[string]interface{}
}

// testCaseColumns is the number of columns the test case CSV schema defines:
// case_id, case_title, step_id, sleep_ms, step_desc, action_type,
// verify_required, test_tool, msg_type, test_data.
const testCaseColumns = 10

// Parse parses CSV data and returns test cases.
func (p *CSVCaseParser) Parse() ([]*TestCase, error) {
	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	_, _ = reader.Read() // skip header

	var cases []*TestCase
	var currentCase *TestCase

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < testCaseColumns {
			return nil, fmt.Errorf("invalid row (got %d columns, expected %d): %v", len(record), testCaseColumns, record)
		}

		if strings.TrimSpace(record[0]) != "" {
			currentCase = &TestCase{
				CaseID:    record[0],
				CaseTitle: record[1],
				Steps:     []TestStep{},
			}
			cases = append(cases, currentCase)
		}

		if currentCase == nil {
			continue
		}

		step := TestStep{
			StepID:         record[2],
			SleepMs:        record[3],
			StepDesc:       record[4],
			ActionType:     record[5],
			VerifyRequired: strings.EqualFold(strings.TrimSpace(record[6]), "Y"),
			TestTool:       record[7],
			MsgType:        record[8],
			TestData:       record[9],
		}
		data, err := p.findTestData(step.TestData, step.StepID)
		if err != nil {
			log.Warnf("Skipping step %s: cannot load test data %s: %v", step.StepID, step.TestData, err)
			continue
		}
		step.TestDatas = data
		currentCase.Steps = append(currentCase.Steps, step)
	}
	return cases, nil
}

func (p *CSVCaseParser) findTestData(sheetName, stepID string) (map[string]interface{}, error) {
	if p.sheetCache == nil {
		p.sheetCache = make(map[string]map[string]map[string]interface{})
	}
	dataFile := filepath.Join(filepath.Dir(p.FilePath), sheetName+filepath.Ext(p.FilePath))
	if _, loaded := p.sheetCache[sheetName]; !loaded {
		data, err := LoadCSVToMap(dataFile)
		if err != nil {
			return nil, err
		}
		p.sheetCache[sheetName] = data
	}
	record, ok := p.sheetCache[sheetName][stepID]
	if !ok {
		return nil, fmt.Errorf("step %s not found in test data file %s", stepID, dataFile)
	}
	return record, nil
}
