package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/google/go-cmp/cmp"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

// Diff represents a difference between two structs.
// It contains the path to the field, the expected value, and the actual value.
type Diff struct {
	Path   string
	Expect interface{}
	Actual interface{}
}

// CompareResult holds the result of the comparison.
// It contains a boolean indicating if the structs are equal and a slice of differences.
type CompareResult struct {
	Equal    bool
	Diffs    []Diff
	DiffInfo string
}

// DiffReporter is a custom reporter for cmp.Diff that collects differences.
type DiffReporter struct {
	path  cmp.Path
	diffs []Diff
}

// PushStep adds a step to the path.
func (r *DiffReporter) PushStep(ps cmp.PathStep) { r.path = append(r.path, ps) }

// PopStep removes the last step from the path.
func (r *DiffReporter) PopStep() { r.path = r.path[:len(r.path)-1] }

// Report is called for each difference found by cmp.Diff.
func (r *DiffReporter) Report(result cmp.Result) {
	if !result.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, Diff{
			r.path.String(),
			formatValue(vx),
			formatValue(vy),
		})
	}
}

func formatValue(v reflect.Value) interface{} {
	if !v.IsValid() {
		return "<nil>"
	}
	return v.Interface()
}

// CompareStruct compares two structs and returns a CompareResult.
func CompareStruct(a, b interface{}) CompareResult {
	r := &DiffReporter{}
	cmp.Diff(a, b, cmp.Reporter(r))
	return CompareResult{
		Equal: len(r.diffs) == 0,
		Diffs: r.diffs,
	}
}

// CompareJSON compares two JSON-like structures (maps and slices) and returns a list of differences.
func CompareJSON(expect, actual interface{}) (CompareResult, error) {
	var eMap, aMap map[string]interface{}
	// Step 1: Marshal expect & actual
	eBytes, err := json.Marshal(expect)
	if err != nil {
		return CompareResult{}, fmt.Errorf("marshal expect failed: %w", err)
	}
	aBytes, err := json.Marshal(actual)
	if err != nil {
		return CompareResult{}, fmt.Errorf("marshal actual failed: %w", err)
	}

	// Step 2: Unmarshal to map[string]interface{}
	if err := json.Unmarshal(eBytes, &eMap); err != nil {
		return CompareResult{}, fmt.Errorf("unmarshal expect failed: %w", err)
	}
	if err := json.Unmarshal(aBytes, &aMap); err != nil {
		return CompareResult{}, fmt.Errorf("unmarshal actual failed: %w", err)
	}

	// Step 3: Compare the two JSON objects recursively
	var diffs []Diff
	compareJSON("", eMap, aMap, &diffs)
	return CompareResult{
		Equal: len(diffs) == 0,
		Diffs: diffs,
	}, nil
}

// compareJSON recursively compares JSON structures.
func compareJSON(path string, expect, actual interface{}, diffs *[]Diff) {
	switch expectType := expect.(type) {
	case map[string]interface{}:
		actualType, ok := actual.(map[string]interface{})
		if !ok {
			*diffs = append(*diffs, Diff{Path: path, Expect: expect, Actual: actual})
			return
		}
		for k, v := range expectType {
			newPath := path + "." + k
			compareJSON(newPath, v, actualType[k], diffs)
		}
		for k := range actualType {
			if _, exists := expectType[k]; !exists {
				newPath := path + "." + k
				compareJSON(newPath, nil, actualType[k], diffs)
			}
		}
	case []interface{}:
		bVal, ok := actual.([]interface{})
		if !ok || len(expectType) != len(bVal) {
			*diffs = append(*diffs, Diff{Path: path, Expect: expect, Actual: actual})
			return
		}
		for i := range expectType {
			compareJSON(path+fmt.Sprintf("[%d]", i), expectType[i], bVal[i], diffs)
		}
	default:
		if expect != actual {
			*diffs = append(*diffs, Diff{Path: path, Expect: expect, Actual: actual})
		}
	}
}

// PrintCompareResult prints the comparison result in a table format.
func PrintCompareResult(result CompareResult) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Path", "Expected", "Actual"})

	for _, diff := range result.Diffs {
		table.Append([]string{
			diff.Path,
			fmt.Sprintf("%v", diff.Expect),
			fmt.Sprintf("%v", diff.Actual),
		})
	}

	if result.Equal {
		log.Info("\n✅ Pass.")
	} else {
		log.Error("\n❌ Diff:")
		table.Render()
	}
}
