package testcase

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// writeCSV is a test helper writing a CSV file into dir.
func writeCSV(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

// TestCrossFileNestedReferences verifies a CSV data sheet can nest another
// file's content via the @file reference: "@name" is an object (one row),
// "@name[]" an array of all rows.
func TestCrossFileNestedReferences(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, dir, "case.csv",
		"case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n"+
			"c1,title,s1,1,d,Send,N,tool,100101,order\n")
	writeCSV(t, dir, "order.csv",
		"StepId,ClOrdID,Partition,Single,SingleAsArray\n"+
			"s1,c0001,@parts[],@single,@single[]\n")
	writeCSV(t, dir, "parts.csv",
		"StepId,PlatformID\n"+
			"p1,101\n"+
			"p2,202\n")
	writeCSV(t, dir, "single.csv",
		"StepId,StopPx\n"+
			"only,999\n")

	cases, err := (&CSVCaseParser{FilePath: filepath.Join(dir, "case.csv")}).Parse()
	assert.NoError(t, err)
	assert.Len(t, cases, 1)
	data := cases[0].Steps[0].TestDatas
	assert.Equal(t, "c0001", data["ClOrdID"])

	partition, ok := data["Partition"].([]map[string]interface{})
	assert.True(t, ok, "array reference should expand to an array, got %T", data["Partition"])
	assert.Len(t, partition, 2)
	assert.Equal(t, "101", partition[0]["PlatformID"])
	assert.Equal(t, "202", partition[1]["PlatformID"])

	single, ok := data["Single"].(map[string]interface{})
	assert.True(t, ok, "object reference should expand to an object, got %T", data["Single"])
	assert.Equal(t, "999", single["StopPx"])

	asArray, ok := data["SingleAsArray"].([]map[string]interface{})
	assert.True(t, ok, "array reference to a one-row sheet should be a single-element array, got %T", data["SingleAsArray"])
	assert.Len(t, asArray, 1)
}

// TestCrossFileNestedReferenceShapeErrors verifies the object reference
// rejects multi-row and empty sheets with guidance, instead of guessing.
func TestCrossFileNestedReferenceShapeErrors(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, dir, "case.csv",
		"case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n"+
			"c1,title,s_multi,1,d,Send,N,tool,100101,order_multi\n"+
			",,s_empty,1,d,Send,N,tool,100101,order_empty\n")
	writeCSV(t, dir, "order_multi.csv", "StepId,Multi\ns_multi,@multi\n")
	writeCSV(t, dir, "order_empty.csv", "StepId,Empty\ns_empty,@empty\n")
	writeCSV(t, dir, "multi.csv", "StepId,X\nr1,1\nr2,2\n")
	writeCSV(t, dir, "empty.csv", "StepId,X\n")

	cases, err := (&CSVCaseParser{FilePath: filepath.Join(dir, "case.csv")}).Parse()
	assert.NoError(t, err)
	steps := cases[0].Steps
	assert.Len(t, steps, 2)
	assert.Contains(t, steps[0].SkipReason, `referenced test data sheet "multi" has 2 rows`)
	assert.Contains(t, steps[0].SkipReason, `use "@multi[]" for an array`)
	assert.Contains(t, steps[1].SkipReason, `referenced test data sheet "empty" is empty`)
}

// TestCrossFileNestedReferenceCycle verifies circular references fail the
// step instead of looping forever.
func TestCrossFileNestedReferenceCycle(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, dir, "case.csv",
		"case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n"+
			"c1,title,s1,1,d,Send,N,tool,100101,a\n")
	writeCSV(t, dir, "a.csv", "StepId,X\ns1,@b\n")
	writeCSV(t, dir, "b.csv", "StepId,Y\ns1,@a\n")

	cases, err := (&CSVCaseParser{FilePath: filepath.Join(dir, "case.csv")}).Parse()
	assert.NoError(t, err)
	assert.Len(t, cases[0].Steps, 1)
	assert.Contains(t, cases[0].Steps[0].SkipReason, "circular test data reference")
}

// TestCrossFileNestedReferenceMissing verifies a reference to a missing
// file/sheet keeps the step with a SkipReason.
func TestCrossFileNestedReferenceMissing(t *testing.T) {
	dir := t.TempDir()
	writeCSV(t, dir, "case.csv",
		"case_id,case_title,step_id,sleep_ms,step_desc,action_type,verify_required,test_tool,msg_type,test_data\n"+
			"c1,title,s1,1,d,Send,N,tool,100101,order\n")
	writeCSV(t, dir, "order.csv", "StepId,X\ns1,@no_such\n")

	cases, err := (&CSVCaseParser{FilePath: filepath.Join(dir, "case.csv")}).Parse()
	assert.NoError(t, err)
	assert.Contains(t, cases[0].Steps[0].SkipReason, "no_such")
}

// TestCrossSheetNestedReferences verifies the same @ reference works across
// sheets inside an Excel workbook.
func TestCrossSheetNestedReferences(t *testing.T) {
	f := excelize.NewFile()
	caseSheet := f.GetSheetName(0)
	caseRows := [][]any{
		{"case_id", "case_title", "step_id", "sleep_ms", "step_desc", "action_type", "verify_required", "test_tool", "msg_type", "test_data"},
		{"c1", "title", "s1", 1, "d", "Send", "N", "tool", "100101", "order"},
	}
	for i, row := range caseRows {
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		assert.NoError(t, f.SetSheetRow(caseSheet, cell, &row))
	}
	_, _ = f.NewSheet("order")
	assert.NoError(t, f.SetSheetRow("order", "A1", &[]any{"StepId", "ClOrdID", "Partition"}))
	assert.NoError(t, f.SetSheetRow("order", "A2", &[]any{"s1", "c0001", "@parts[]"}))
	_, err := f.NewSheet("parts")
	assert.NoError(t, err)
	assert.NoError(t, f.SetSheetRow("parts", "A1", &[]any{"StepId", "PlatformID"}))
	assert.NoError(t, f.SetSheetRow("parts", "A2", &[]any{"p1", 101}))
	assert.NoError(t, f.SetSheetRow("parts", "A3", &[]any{"p2", 202}))

	path := filepath.Join(t.TempDir(), "case.xlsx")
	assert.NoError(t, f.SaveAs(path))

	cases, err := (&ExcelCaseParser{FilePath: path}).Parse()
	assert.NoError(t, err)
	data := cases[0].Steps[0].TestDatas
	partition, ok := data["Partition"].([]map[string]interface{})
	assert.True(t, ok, "cross-sheet reference should expand to an array, got %T", data["Partition"])
	assert.Len(t, partition, 2)
	assert.Equal(t, "101", partition[0]["PlatformID"])
	assert.Equal(t, "202", partition[1]["PlatformID"])
}
