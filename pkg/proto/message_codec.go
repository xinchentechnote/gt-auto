package proto

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

// ConvertMapToStruct converts a map to a struct.
func ConvertMapToStruct(data map[string]interface{}, target interface{}) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 使用 JSON tag，如果有的话
		jsonKey := field.Tag.Get("json")
		if jsonKey == "" {
			jsonKey = field.Name
		}

		raw, ok := data[jsonKey]
		if !ok {
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		converted, err := convertValue(raw, field.Type)
		if err != nil {
			return fmt.Errorf("field '%s': %w", field.Name, err)
		}
		fieldValue.Set(converted)
	}

	return nil
}

func convertValue(input interface{}, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		switch v := input.(type) {
		case string:
			return reflect.ValueOf(v), nil
		case float64:
			return reflect.ValueOf(strconv.FormatFloat(v, 'f', -1, 64)), nil
		case int:
			return reflect.ValueOf(strconv.Itoa(v)), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to string", input)
		}
	case reflect.Int, reflect.Int32, reflect.Int64:
		var i int64
		switch v := input.(type) {
		case float64:
			i = int64(v)
		case string:
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			i = n
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to int", input)
		}
		return reflect.ValueOf(i).Convert(targetType), nil

	case reflect.Uint32:
		switch v := input.(type) {
		case float64:
			return reflect.ValueOf(uint32(v)), nil
		case string:
			n, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(uint32(n)), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to uint32", input)
		}
	case reflect.Uint64:
		switch v := input.(type) {
		case float64:
			return reflect.ValueOf(uint64(v)), nil
		case string:
			n, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(uint64(n)), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to uint32", input)
		}

	case reflect.Bool:
		switch v := input.(type) {
		case bool:
			return reflect.ValueOf(v), nil
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(b), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", input)
		}

	case reflect.Struct:
		// 支持嵌套 struct（递归）
		if m, ok := input.(map[string]interface{}); ok {
			val := reflect.New(targetType).Elem()
			err := ConvertMapToStruct(m, val.Addr().Interface())
			return val, err
		}
	case reflect.Uint8:
		switch v := input.(type) {
		case float64:
			return reflect.ValueOf(uint8(v)), nil
		case string:
			n, err := strconv.ParseUint(v, 10, 8)
			if err != nil {
				return reflect.Value{}, err
			}
			return reflect.ValueOf(uint8(n)), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to uint8", input)
		}
	}

	// 默认处理为 JSON 解码
	bytes, err := json.Marshal(input)
	if err != nil {
		return reflect.Value{}, err
	}
	val := reflect.New(targetType).Interface()
	if err := json.Unmarshal(bytes, val); err != nil {
		return reflect.Value{}, err
	}
	return reflect.ValueOf(val).Elem(), nil
}
