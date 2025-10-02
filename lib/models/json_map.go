package models

import (
	"database/sql/driver"
	"discord-werewolf/lib"
	"encoding/json"
)

type JsonMap map[string]interface{}

func (m JsonMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}
func (m *JsonMap) Scan(value interface{}) error {
	return lib.UnMarshalBytes(m, value)
}

func (m JsonMap) GormDataType() string {
	return "json_map"
}
