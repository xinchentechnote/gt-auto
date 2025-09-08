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
func (p *CSVCaseParser) Parse() ([]*TestCase, error) {
	file, err := os.Open(p.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
			log.Infof("Error finding test data for %s step %s: %v\n", step.TestData, step.StepID, err)
			continue
		}
		step.TestDatas = data
		currentCase.Steps = append(currentCase.Steps, step)
		cases[len(cases)-1] = currentCase
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
