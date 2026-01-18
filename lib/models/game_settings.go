package models

import (
	"database/sql/driver"
	"encoding/json"
)

type GameSettings struct {
	GameMode         string                      `json:"gameMode"`
	GameModeSettings map[string]GameModeSettings `json:"gameModeSettings"`
}

func (m GameSettings) Value() (driver.Value, error) {
	return json.Marshal(m)
}
func (m *GameSettings) Scan(value interface{}) error {
	if value == nil {
		m.GameModeSettings = map[string]GameModeSettings{}
		return nil
	}
	return UnmarshalBytes(m, value)
}
func (m GameSettings) GormDataType() string {
	return "game_settings"
}

type GameModeSettings struct {
	Teams      []string            `json:"teams"`
	Characters []CharacterSettings `json:"characters"`
}

type CharacterSettings struct {
	Id    string `json:"id"`
	Count int    `json:"count"`
}
