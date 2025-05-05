package testcase

import (
	"encoding/csv"
	"io"
	"os"
)

// LoadCSVToMap loads a CSV file and returns a map where the keys are the values in the first column
func LoadCSVToMap(filePath string) (map[string]map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// 读取表头
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	records := make(map[string]map[string]interface{})
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		record := make(map[string]interface{})
		for i, header := range headers {
			value := row[i]

			// 尝试转为 int 类型（如果失败则保留为字符串）
			// if intVal, err := strconv.Atoi(value); err == nil {
			// 	record[header] = intVal
			// } else {
			record[header] = value
			// }
		}
		records[record["StepId"].(string)] = record
	}

	return records, nil
}
