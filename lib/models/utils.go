package models

import (
	"encoding/json"
	"fmt"
)

func UnmarshalBytes[T any](m *T, value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to unmarshal JSONB from %v", value)
	}
	return json.Unmarshal(bytes, m)
}
