package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JsonArray[T any] []T

func (ja *JsonArray[T]) Value() (driver.Value, error) {
	return json.Marshal(ja)
}

func (ja *JsonArray[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to unmarshal JSONB from %v", value)
	}
	return json.Unmarshal(bytes, &ja)
}
