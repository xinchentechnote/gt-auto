package parser

import (
	"encoding/csv"
	"io"
	"os"
	"strings"
)

// TestStep represents a single step in a test case.
type TestStep struct {
	StepID        string
	SleepTime     string
	Desc          string
	ActionType    string
	TestTool      string
	TestFunction  string
	TestDataSheet string
}

// TestCase represents a test case with its steps.
type TestCase struct {
	CaseNo    string
	CaseTitle string
	Steps     []TestStep
}

// CaseParser is an interface for parsing test cases from different formats.
type CaseParser interface {
	Parse() ([]TestCase, error)
}

// CSVCaseParser implements the CaseParser interface for CSV files.
type CSVCaseParser struct {
	FilePath string
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
		currentCase.Steps = append(currentCase.Steps, step)
		cases[len(cases)-1] = *currentCase
	}
	return cases, nil
}
