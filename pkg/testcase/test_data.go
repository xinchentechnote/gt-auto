package testcase

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadCSVToMap loads a CSV file and returns a map where the keys are the values in the first column
func LoadCSVToMap(filePath string) (map[string]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	var rows [][]string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	return recordsFromRows(rows, filePath)
}

// recordsFromRows converts sheet rows (header row first) into records keyed
// by the StepId column. Rows shorter than the header are padded with empty
// strings - spreadsheets trim trailing empty cells - and fully empty rows are
// skipped.
func recordsFromRows(rows [][]string, source string) (map[string]map[string]interface{}, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s has no header row", source)
	}
	headers := rows[0]

	records := make(map[string]map[string]interface{})
	for _, row := range rows[1:] {
		if isEmptyRow(row) {
			continue
		}

		record := make(map[string]interface{})
		for i, header := range headers {
			value := ""
			if i < len(row) {
				value = row[i]
			}

			// 尝试转为 int 类型（如果失败则保留为字符串）
			// if intVal, err := strconv.Atoi(value); err == nil {
			// 	record[header] = intVal
			// } else {
			record[header] = value
			// }
		}
		stepID, ok := record["StepId"].(string)
		if !ok {
			return nil, fmt.Errorf("%s has no StepId column", source)
		}
		records[stepID] = record
	}

	return records, nil
}

// isEmptyRow reports whether every cell in the row is empty.
func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
