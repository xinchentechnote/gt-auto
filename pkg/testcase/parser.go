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

// testCaseColumns is the number of columns the test case CSV schema defines:
// case_id, case_title, step_id, sleep_ms, step_desc, action_type,
// verify_required, test_tool, msg_type, test_data.
const testCaseColumns = 10

// sheetLoader loads one test data sheet by name, returning its records keyed
// by StepId.
type sheetLoader func(sheetName string) (map[string]map[string]interface{}, error)

// sheetCache loads every sheet at most once and keeps records isolated per
// sheet so identical StepIds in different sheets cannot collide.
type sheetCache map[string]map[string]map[string]interface{}

func (c sheetCache) lookup(sheetName, stepID string, load sheetLoader) (map[string]interface{}, error) {
	if _, loaded := c[sheetName]; !loaded {
		data, err := load(sheetName)
		if err != nil {
			return nil, err
		}
		c[sheetName] = data
	}
	record, ok := c[sheetName][stepID]
	if !ok {
		return nil, fmt.Errorf("step %s not found in test data sheet %s", stepID, sheetName)
	}
	return record, nil
}

// parseCaseRows turns case-table rows (header excluded) into test cases.
// Rows follow the testCaseColumns schema; a non-empty first column starts a
// new case. Steps whose test data cannot be resolved are kept with a
// SkipReason so the executor records a failure instead of dropping them.
func parseCaseRows(rows [][]string, loadSheet sheetLoader) ([]*TestCase, error) {
	var cases []*TestCase
	var currentCase *TestCase
	cache := sheetCache{}

	for _, record := range rows {
		if isEmptyRow(record) {
			continue
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
		data, err := cache.lookup(step.TestData, step.StepID, loadSheet)
		if err != nil {
			// Keep the step but mark it so the executor records a failure -
			// dropping it here would let the case pass with missing steps.
			log.Warnf("Step %s will fail: cannot load test data %s: %v", step.StepID, step.TestData, err)
			step.SkipReason = err.Error()
		} else {
			step.TestDatas = data
		}
		currentCase.Steps = append(currentCase.Steps, step)
	}
	return cases, nil
}

// CSVCaseParser implements the CaseParser interface for CSV files.
type CSVCaseParser struct {
	FilePath string
}

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

	var rows [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, record)
	}

	return parseCaseRows(rows, p.loadDataSheet)
}

// loadDataSheet reads one test data sheet file (<sheet><ext> next to the
// case file) into records keyed by StepId.
func (p *CSVCaseParser) loadDataSheet(sheetName string) (map[string]map[string]interface{}, error) {
	dataFile := filepath.Join(filepath.Dir(p.FilePath), sheetName+filepath.Ext(p.FilePath))
	return LoadCSVToMap(dataFile)
}
