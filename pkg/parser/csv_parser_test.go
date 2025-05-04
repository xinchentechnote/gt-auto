package parser

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSVCaseParserParse(t *testing.T) {
	filePath := filepath.Join("testdata", "test_case.csv")

	parser := &CSVCaseParser{FilePath: filePath}
	cases, err := parser.Parse()

	assert.NoError(t, err)
	assert.Len(t, cases, 1, "should parse 1 test case")

	tc := cases[0]
	assert.Equal(t, "sz_001", tc.CaseNo)
	assert.Equal(t, "order", tc.CaseTitle)
	assert.Len(t, tc.Steps, 4, "should have 4 steps")

	assert.Equal(t, "new_order_001", tc.Steps[0].StepId)
	assert.Equal(t, "oms1", tc.Steps[0].TestTool)
	assert.Equal(t, "100101", tc.Steps[0].TestFunction)

	assert.Equal(t, "new_order_002", tc.Steps[1].StepId)
	assert.Equal(t, "tgw1", tc.Steps[1].TestTool)
	assert.Equal(t, "100101", tc.Steps[1].TestFunction)

	assert.Equal(t, "new_order_003", tc.Steps[2].StepId)
	assert.Equal(t, "tgw1", tc.Steps[2].TestTool)
	assert.Equal(t, "200102", tc.Steps[2].TestFunction)
}
