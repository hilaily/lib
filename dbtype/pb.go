package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func NewPBJSONType[T any](val *T) *PBJSONType[T] {
	return &PBJSONType[T]{val: val}
}

type PBJSONType[T any] struct {
	val   *T
	empty string
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *PBJSONType[T]) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) == 0 || string(bytes) == "{}" {
		j.val = new(T)
		return nil
	}

	msg := new(T)
	msgM, ok := any(msg).(proto.Message)
	if !ok {
		return fmt.Errorf("failed to create new proto message instance, maybe T is not a proto message(must not be a pointer)")
	}
	err := protojson.Unmarshal(bytes, msgM)
	if err != nil {
		return fmt.Errorf("failed to unmarshal proto message: %w", err)
	}
	j.val = msg
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j PBJSONType[T]) Value() (driver.Value, error) {
	nilT := new(T)
	if j.val == nil || j.val == nilT {
		return []byte("{}"), nil
	}
	logrus.Debugf("j.val: %T, %#+v", j.val, j.val)
	msgM, ok := any(j.val).(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to create new proto message instance, data: %+v", j.val)
	}
	return protojson.Marshal(msgM)
}

func (j PBJSONType[T]) Data() *T {
	return j.val
}

func NewPBJSONSlice[T any](val []*T) *PBJSONSlice[T] {
	return &PBJSONSlice[T]{val: val}
}

type PBJSONSlice[T any] struct {
	val []*T
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *PBJSONSlice[T]) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}

	if len(bytes) == 0 || string(bytes) == "[]" {
		j.val = []*T{}
		return nil
	}

	// 解析成通用的 JSON 数组
	var jsonArray []json.RawMessage
	if err := json.Unmarshal(bytes, &jsonArray); err != nil {
		return fmt.Errorf("failed to unmarshal JSON array: %w", err)
	}

	// 逐个解析每个元素
	result := make([]*T, len(jsonArray))
	for i, itemBytes := range jsonArray {
		// 使用反射创建新的消息实例
		msg := new(T)
		msgM, ok := any(msg).(proto.Message)
		if !ok {
			return fmt.Errorf("failed to create new proto message instance at index %d", i)
		}
		if err := protojson.Unmarshal(itemBytes, msgM); err != nil {
			return fmt.Errorf("failed to unmarshal proto message at index %d: %w", i, err)
		}
		result[i] = msg
	}

	j.val = result
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j PBJSONSlice[T]) Value() (driver.Value, error) {
	if len(j.val) == 0 {
		return []byte("[]"), nil
	}

	// 将每个 proto message 转换为 JSON
	jsonArray := make([]json.RawMessage, len(j.val))
	for i, msg := range j.val {
		if msg == nil {
			jsonArray[i] = []byte("{}")
			continue
		}
		msgM, ok := any(msg).(proto.Message)
		if !ok {
			return nil, fmt.Errorf("failed to create new proto message instance at index %d", i)
		}
		jsonBytes, err := protojson.Marshal(msgM)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal proto message at index %d: %w", i, err)
		}
		jsonArray[i] = jsonBytes
	}

	// 将 JSON 数组序列化
	result, err := json.Marshal(jsonArray)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal final JSON array: %w", err)
	}

	return result, nil
}

func (j PBJSONSlice[T]) Data() []*T {
	return j.val
}
