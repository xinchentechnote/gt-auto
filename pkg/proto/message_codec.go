package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
)

func init() {
	//当前实现参考了深圳证券交易所的二进制协议，进行了简化，仅为了演示功能实现
	//将来将基于antlr4定义dsl描述完整的深交所二进制协议，并生成对应的代码
	RegisterMessage(1, func() Message { return &Logon{} })
	RegisterMessage(2, func() Message { return &Logout{} })
	RegisterMessage(3, func() Message { return &Heartbeat{} })
	RegisterMessage(4, func() Message { return &BusinessReject{} })
	RegisterMessage(100101, func() Message { return &NewOrder{} })
	RegisterMessage(200102, func() Message { return &ExecutionConfirm{} })
	RegisterMessage(200115, func() Message { return &ExecutionReport{} })
	RegisterMessage(190007, func() Message { return &CancelRequest{} })
	RegisterMessage(290008, func() Message { return &CancelReject{} })
}

// SzseMessageFactory is a factory function that creates a new message instance.
// It returns a Message interface, which is implemented by all message types.
type SzseMessageFactory func() Message

var registry = map[uint32]SzseMessageFactory{}

// RegisterMessage registers a message type with its factory function.
func RegisterMessage(msgType uint32, factory SzseMessageFactory) {
	registry[msgType] = factory
}

// BinarySzseMessageCodec is a codec for encoding and decoding messages.
// It implements the MessageCodec interface for the SZSE binary protocol.
type BinarySzseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (codec *BinarySzseMessageCodec) EncodeJSONMap(message map[string]interface{}) ([]byte, error) {
	data, e := codec.JSONToStruct(message)
	if e != nil {
		return nil, fmt.Errorf("failed to encode message: %w", e)
	}
	return codec.Encode(data)
}

// JSONToStruct implements MessageCodec.
func (codec *BinarySzseMessageCodec) JSONToStruct(jsonMap map[string]interface{}) (Message, error) {
	msgType, err := strconv.Atoi(jsonMap["MsgType"].(string))
	if err != nil {
		return nil, fmt.Errorf("unknown MsgType: %s", jsonMap["MsgType"].(string))
	}
	factory, ok := registry[uint32(msgType)]
	if !ok {
		return nil, fmt.Errorf("unknown MsgType: %s", jsonMap["MsgType"].(string))
	}
	message := factory()
	err = ConvertMapToStruct(jsonMap, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

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

// Encode a message into a byte slice and prepends the message type and length.
func (codec *BinarySzseMessageCodec) Encode(message Message) ([]byte, error) {
	// 将字符串 MsgType 转换为 int32
	msgType := message.MsgType()

	data, err := message.Encode()
	if err != nil {
		return nil, err
	}

	length := len(data)
	b := make([]byte, 8+length)
	binary.BigEndian.PutUint32(b[0:4], uint32(msgType))
	binary.BigEndian.PutUint32(b[4:8], uint32(length))
	copy(b[8:], data)

	return b, nil
}

// Decode a byte slice into a message.
func (codec *BinarySzseMessageCodec) Decode(data []byte) (Message, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	msgTypeInt := binary.BigEndian.Uint32(data[0:4])
	length := binary.BigEndian.Uint32(data[4:8])
	if len(data) < int(8+length) {
		return nil, fmt.Errorf("data length mismatch")
	}

	body := data[8 : 8+length]

	factory, ok := registry[msgTypeInt]
	if !ok {
		return nil, fmt.Errorf("unknown MsgType: %d", msgTypeInt)
	}

	msg := factory()
	if err := msg.Decode(body); err != nil {
		return nil, err
	}

	return msg, nil
}

// BinarySseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE binary protocol.
type BinarySseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (b *BinarySseMessageCodec) EncodeJSONMap(map[string]interface{}) ([]byte, error) {
	panic("unimplemented")
}

// JSONToStruct implements MessageCodec.
func (b *BinarySseMessageCodec) JSONToStruct(map[string]interface{}) (Message, error) {
	panic("unimplemented")
}

// Decode implements MessageCodec.
func (b *BinarySseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (b *BinarySseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}

// StepSzseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SZSE step protocol.
type StepSzseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (s *StepSzseMessageCodec) EncodeJSONMap(map[string]interface{}) ([]byte, error) {
	panic("unimplemented")
}

// JSONToStruct implements MessageCodec.
func (s *StepSzseMessageCodec) JSONToStruct(map[string]interface{}) (Message, error) {
	panic("unimplemented")
}

// Decode implements MessageCodec.
func (s *StepSzseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (s *StepSzseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}

// StepSseMessageCodec is a codec for encoding and decoding messages.
// It is a placeholder and should be implemented according to the SSE step protocol.
type StepSseMessageCodec struct{}

// EncodeJSONMap implements MessageCodec.
func (s *StepSseMessageCodec) EncodeJSONMap(map[string]interface{}) ([]byte, error) {
	panic("unimplemented")
}

// JSONToStruct implements MessageCodec.
func (s *StepSseMessageCodec) JSONToStruct(map[string]interface{}) (Message, error) {
	panic("unimplemented")
}

// Decode implements MessageCodec.
func (s *StepSseMessageCodec) Decode([]byte) (Message, error) {
	panic("unimplemented")
}

// Encode implements MessageCodec.
func (s *StepSseMessageCodec) Encode(Message) ([]byte, error) {
	panic("unimplemented")
}
