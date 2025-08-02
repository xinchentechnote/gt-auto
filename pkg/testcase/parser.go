package testcase

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

// CSVCaseParser implements the CaseParser interface for CSV files.
type CSVCaseParser struct {
	FilePath      string
	testDataCache map[string]map[string]interface{}
}

// Parse parses CSV data and returns test cases.
func (p *CSVCaseParser) Parse() ([]TestCase, error) {
	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	_, _ = reader.Read() // skip header

	var cases []TestCase
	var currentCase *TestCase

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.TrimSpace(record[0]) != "" {
			currentCase = &TestCase{
				CaseNo:    record[0],
				CaseTitle: record[1],
				Steps:     []TestStep{},
			}
			cases = append(cases, *currentCase)
		}

		if currentCase == nil {
			continue
		}

		step := TestStep{
			StepID:        record[2],
			SleepTime:     record[3],
			Desc:          record[4],
			ActionType:    record[5],
			TestTool:      record[6],
			TestFunction:  record[7],
			TestDataSheet: record[8],
		}
		data, err := p.findTestData(step.TestDataSheet, step.StepID)
		if err != nil {
			log.Infof("Error finding test data for %s step %s: %v\n", step.TestDataSheet, step.StepID, err)
			continue
		}
		step.TestData = data
		currentCase.Steps = append(currentCase.Steps, step)
		cases[len(cases)-1] = *currentCase
	}
	return cases, nil
}

func (p *CSVCaseParser) findTestData(sheetName, stepID string) (map[string]interface{}, error) {
	if p.testDataCache == nil {
		p.testDataCache = make(map[string]map[string]interface{})
	}
	if data, ok := p.testDataCache[stepID]; ok {
		return data, nil
	}
	dir := filepath.Dir(p.FilePath)
	ext := filepath.Ext(p.FilePath)
	data, err := LoadCSVToMap(filepath.Join(dir, sheetName+ext))
	if err != nil {
		return nil, err
	}
	for k, v := range data {
		p.testDataCache[k] = v
	}

	return p.testDataCache[stepID], nil
}
