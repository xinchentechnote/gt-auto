package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/olekukonko/tablewriter"
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
	Equal bool
	Diffs []Diff
}

// CompareStruct compares two structs and returns a CompareResult.
func CompareStruct(a, b interface{}) CompareResult {
	var diffs []Diff
	compareStruct(reflect.ValueOf(a), reflect.ValueOf(b), "", &diffs)
	return CompareResult{
		Equal: len(diffs) == 0,
		Diffs: diffs,
	}
}

func compareStruct(expect, actual reflect.Value, path string, diffs *[]Diff) {
	if expect.Kind() == reflect.Ptr {
		expect = expect.Elem()
	}
	if actual.Kind() == reflect.Ptr {
		actual = actual.Elem()
	}

	if !expect.IsValid() || !actual.IsValid() {
		if expect.IsValid() != actual.IsValid() {
			*diffs = append(*diffs, Diff{Path: path, Expect: fmtVal(expect), Actual: fmtVal(actual)})
		}
		return
	}

	if expect.Type() != actual.Type() {
		*diffs = append(*diffs, Diff{Path: path, Expect: fmtVal(expect), Actual: fmtVal(actual)})
		return
	}

	switch expect.Kind() {
	case reflect.Struct:
		for i := 0; i < expect.NumField(); i++ {
			field := expect.Type().Field(i)
			if field.PkgPath != "" {
				continue
			}
			fieldPath := field.Name
			if path != "" {
				fieldPath = path + "." + field.Name
			}
			compareStruct(expect.Field(i), actual.Field(i), fieldPath, diffs)
		}
	case reflect.Slice, reflect.Array:
		if expect.Len() != actual.Len() {
			*diffs = append(*diffs, Diff{Path: path + ".length", Expect: expect.Len(), Actual: actual.Len()})
			return
		}
		for i := 0; i < expect.Len(); i++ {
			elemPath := fmt.Sprintf("%s[%d]", path, i)
			compareStruct(expect.Index(i), actual.Index(i), elemPath, diffs)
		}
	default:
		if !reflect.DeepEqual(expect.Interface(), actual.Interface()) {
			*diffs = append(*diffs, Diff{
				Path:   path,
				Expect: expect.Interface(),
				Actual: actual.Interface(),
			})
		}
	}
}

func fmtVal(v reflect.Value) interface{} {
	if !v.IsValid() {
		return nil
	}
	return v.Interface()
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
		fmt.Println("✅ Pass.")
	} else {
		fmt.Println("❌ Diff:")
		table.Render()
	}
}
