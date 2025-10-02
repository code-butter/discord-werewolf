package models

import (
	"database/sql/driver"
	"discord-werewolf/lib"
	"encoding/json"
	"time"
)

type Guild struct {
	Id           string `gorm:"primary_key"`
	Name         string
	Channels     GuildChannels
	Paused       bool
	GameGoing    bool
	DayNight     bool
	TimeZone     string
	DayTime      *TimeOnly
	NightTime    *TimeOnly
	GameSettings JsonMap
	LastCycleRan *time.Time `gorm:"type:datetime"`
}
type GuildChannel struct {
	Id       string
	Name     string
	AppId    string
	Children *[]GuildChannel
}

type GuildChannels map[string]GuildChannel

func (m GuildChannels) Value() (driver.Value, error) {
	return json.Marshal(m)
}
func (m *GuildChannels) Scan(value interface{}) error {
	return lib.UnMarshalBytes(m, value)
}

func (m GuildChannels) GormDataType() string {
	return "guild_channels"
}
