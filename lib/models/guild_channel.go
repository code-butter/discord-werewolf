package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Used for AppIds and initial channel names

const ChannelTownSquare = "town-square"
const ChannelWerewolves = "werewolves"
const ChannelWitch = "witch"
const ChannelMasons = "masons"
const ChannelBodyguard = "bodyguard"
const ChannelAfterLife = "after-life"

const CatChannelTheTown = "the-town"
const CatChannelInstructions = "game-instructions"
const CatChannelAdmin = "admin"

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
	return UnmarshalBytes(m, value)
}

func (m GuildChannels) GormDataType() string {
	return "guild_channels"
}
