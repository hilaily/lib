package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

func NewJSONType[T any](val *T) *JSONType[T] {
	return &JSONType[T]{val: val}
}

type JSONType[T any] struct {
	val   *T
	empty string
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *JSONType[T]) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	if len(bytes) == 0 {
		j.val = new(T)
		return nil
	}

	result := new(T)
	err := json.Unmarshal(bytes, result)
	j.val = result
	return err
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j JSONType[T]) Value() (driver.Value, error) {
	if j.val == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(j.val)
}

func (j JSONType[T]) Data() T {
	return *j.val
}

func NewJSONSlice[T any](val []T) JSONSlice[T] {
	return JSONSlice[T]{val: val}
}

type JSONSlice[T any] struct {
	val []T
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *JSONSlice[T]) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSON value:", value))
	}
	result := []T{}
	if len(bytes) == 0 || string(bytes) == "[]" {
		j.val = []T{}
		return nil
	}

	err := json.Unmarshal(bytes, &result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON value, value: %v, err: %w", value, err)
	}
	j.val = result
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j JSONSlice[T]) Value() (driver.Value, error) {
	if len(j.val) == 0 {
		return []byte("[]"), nil
	}
	en, err := json.Marshal(j.val)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON value, value: %v, err: %w", j.val, err)
	}
	logrus.Debugf("en: %+v\n", string(en))
	return en, nil
}

func (j JSONSlice[T]) Data() []T {
	return j.val
}
