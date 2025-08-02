package validate

import (
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
