package models

import (
	"database/sql/driver"
	"discord-werewolf/lib"
	"encoding/json"
)

// Used for AppIds and initial channel names

const ChannelTownSquare = "town-square"
const ChannelWerewolves = "werewolves"
const ChannelWitch = "witch"
const ChannelMasons = "masons"
const ChannelBodyguard = "bodyguard"
const ChannelAfterLife = "after-life"

// Used for AppIds on channels with possibly multiple

const ChannelSeerPrefix = "seer-"
const ChannelLoversPrefix = "lovers-"

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
