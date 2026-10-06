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

// refPrefix marks a data sheet cell whose value references another data
// sheet (Excel) or file (CSV). The referenced content becomes the field's
// nested value: a single object for a one-row sheet, an array of objects for
// multiple rows. See docs/design.md for the format spec.
const refPrefix = "@"

// sheetLoader loads one test data sheet by name, preserving row order.
type sheetLoader func(sheetName string) (*sheetData, error)

// sheetCache loads every sheet at most once, resolves cross-sheet @
// references (with cycle detection), and keeps sheets isolated so identical
// StepIds in different sheets cannot collide.
type sheetCache struct {
	sheets map[string]*sheetData
	load   sheetLoader
}

func newSheetCache(load sheetLoader) *sheetCache {
	return &sheetCache{sheets: make(map[string]*sheetData), load: load}
}

// lookup returns the record of one step from the named sheet.
func (c *sheetCache) lookup(sheetName, stepID string) (map[string]interface{}, error) {
	sd, err := c.get(sheetName, nil)
	if err != nil {
		return nil, err
	}
	record, ok := sd.byStep[stepID]
	if !ok {
		return nil, fmt.Errorf("step %s not found in test data sheet %s", stepID, sheetName)
	}
	return record, nil
}

// get loads and fully expands a sheet; chain carries the sheet names currently
// being expanded for cycle detection.
func (c *sheetCache) get(sheetName string, chain []string) (*sheetData, error) {
	for _, name := range chain {
		if name == sheetName {
			cycle := strings.Join(append(append([]string{}, chain...), sheetName), " -> ")
			return nil, fmt.Errorf("circular test data reference: %s", cycle)
		}
	}
	if sd, ok := c.sheets[sheetName]; ok {
		return sd, nil
	}
	raw, err := c.load(sheetName)
	if err != nil {
		return nil, err
	}
	sd, err := c.expandSheet(raw, sheetName, chain)
	if err != nil {
		return nil, err
	}
	c.sheets[sheetName] = sd
	return sd, nil
}

// expandSheet replaces every "@sheet" cell value with the referenced sheet's
// content: a single object for a one-row sheet, an array of objects for
// multiple rows. Referenced sheets are expanded recursively before embedding.
func (c *sheetCache) expandSheet(sd *sheetData, sheetName string, chain []string) (*sheetData, error) {
	next := append(append([]string{}, chain...), sheetName)
	for _, record := range sd.rows {
		for key, value := range record {
			ref, ok := value.(string)
			if !ok || !strings.HasPrefix(ref, refPrefix) {
				continue
			}
			refData, err := c.get(strings.TrimPrefix(ref, refPrefix), next)
			if err != nil {
				return nil, err
			}
			if len(refData.rows) == 1 {
				record[key] = refData.rows[0]
			} else {
				record[key] = refData.rows
			}
		}
	}
	return sd, nil
}

// parseCaseRows turns case-table rows (header excluded) into test cases.
// Rows follow the testCaseColumns schema; a non-empty first column starts a
// new case. Steps whose test data cannot be resolved are kept with a
// SkipReason so the executor records a failure instead of dropping them.
func parseCaseRows(rows [][]string, loadSheet sheetLoader) ([]*TestCase, error) {
	var cases []*TestCase
	var currentCase *TestCase
	cache := newSheetCache(loadSheet)

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
		data, err := cache.lookup(step.TestData, step.StepID)
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
// case file) preserving row order.
func (p *CSVCaseParser) loadDataSheet(sheetName string) (*sheetData, error) {
	return loadCSVSheet(filepath.Join(filepath.Dir(p.FilePath), sheetName+filepath.Ext(p.FilePath)))
}
