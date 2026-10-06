package testcase

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// sheetData holds one loaded data sheet: rows in sheet order plus a StepId
// index into them.
type sheetData struct {
	rows   []map[string]interface{}
	byStep map[string]map[string]interface{}
}

// LoadCSVToMap loads a CSV data file into records keyed by the StepId column.
func LoadCSVToMap(filePath string) (map[string]map[string]interface{}, error) {
	sd, err := loadCSVSheet(filePath)
	if err != nil {
		return nil, err
	}
	return sd.byStep, nil
}

// loadCSVSheet loads a CSV data file preserving the row order.
func loadCSVSheet(filePath string) (*sheetData, error) {
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

	return sheetFromRows(rows, filePath)
}

// sheetFromRows converts sheet rows (header row first) into a sheetData.
// Rows shorter than the header are padded with empty strings - spreadsheets
// trim trailing empty cells - and fully empty rows are skipped.
func sheetFromRows(rows [][]string, source string) (*sheetData, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s has no header row", source)
	}
	headers := rows[0]

	sd := &sheetData{byStep: make(map[string]map[string]interface{})}
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
		sd.rows = append(sd.rows, record)
		sd.byStep[stepID] = record
	}

	return sd, nil
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
