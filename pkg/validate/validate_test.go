package validate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	Name    string
	Age     int
	City    string
	Address *Address
}

func TestCompareStruct(t *testing.T) {
	tests := []struct {
		name    string
		expect  interface{}
		actual  interface{}
		equal   bool // whether the structs are expected to be equal
		diffLen int  // expected number of diffs
	}{
		{
			name:    "equal structs",
			expect:  Person{Name: "Alice", Age: 30, City: "New York"},
			actual:  Person{Name: "Alice", Age: 30, City: "New York"},
			equal:   true,
			diffLen: 0,
		},
		{
			name:    "different values",
			expect:  Person{Name: "Alice", Age: 30, City: "New York"},
			actual:  Person{Name: "Alice", Age: 31, City: "Boston"},
			equal:   false,
			diffLen: 2,
		},
		{
			name:    "different types",
			expect:  Person{Name: "Alice", Age: 30, City: "New York"},
			actual:  struct{ Name string }{"Alice"},
			equal:   false,
			diffLen: 1, // type mismatch
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareStruct(tt.expect, tt.actual)
			assert.Equal(t, tt.equal, result.Equal)
			assert.Len(t, result.Diffs, tt.diffLen)
			PrintCompareResult(result)
		})
	}
}

type Address struct {
	City    string
	Country string
}

func TestCompareJSON(t *testing.T) {
	tests := []struct {
		name    string
		expect  interface{}
		actual  interface{}
		equal   bool // expected equality
		diffLen int  // expected number of diffs
	}{
		{
			name: "identical JSON objects",
			expect: Person{
				Name:    "Alice",
				Age:     30,
				City:    "New York",
				Address: &Address{City: "Shanghai", Country: "China"},
			},
			actual: Person{
				Name:    "Alice",
				Age:     30,
				City:    "New York",
				Address: &Address{City: "Shanghai", Country: "China"},
			},
			equal:   true,
			diffLen: 0,
		},
		{
			name: "different values",
			expect: Person{
				Name:    "Alice",
				Age:     30,
				City:    "New York",
				Address: &Address{City: "Shanghai", Country: "China"},
			},
			actual: Person{
				Name:    "Alice",
				Age:     31,
				City:    "Boston",
				Address: &Address{City: "Beijing", Country: "China"},
			},
			equal:   false,
			diffLen: 3, // 3 diffs expected: Name, Age, and Address
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := CompareJSON(tt.expect, tt.actual)
			assert.Equal(t, tt.equal, result.Equal)
			assert.Len(t, result.Diffs, tt.diffLen)
			PrintCompareResult(result)
		})
	}
}
