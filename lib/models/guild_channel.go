package models

import (
	"database/sql/driver"
	"encoding/json"
)

// Town channels

const ChannelTownSquare = "town-square"
const ChannelWerewolves = "werewolves"
const ChannelWitch = "witch"
const ChannelMasons = "masons"
const ChannelBodyguard = "bodyguard"
const ChannelAfterLife = "after-life"

// Admin channels

const ChannelAdminSettings = "admin-settings"
const ChannelAdminCharacters = "admin-characters"

// Instruction channels

const ChannelHowToPlay = "how-to-play"
const ChannelPlayerRoles = "player-roles"
const ChannelCommands = "commands"

// Parent channels

const CatChannelTheTown = "the-town"
const CatChannelInstructions = "game-instructions"
const CatChannelAdmin = "admin"

// Used for AppIds on channels with possibly multiple

const ChannelSeerPrefix = "seer-"
const ChannelLoversPrefix = "lovers-"

// Top level message actions on config channels

const MessageGameMode = "game-mode"
const MessageGameMode_BalancedTeams = "balanced-teams"
const MessageGameMode_BalancedSelect = "balanced-select"
const MessageGameMode_RandomTeams = "random-teams"
const MessageGameMode_RandomSelect = "random-select"

const MessageTeam = "team"
const MessageCharacterCount = "character-count"
const MessageCharacterToggle = "character-toggle"

const ModalCharacterCount = "character-count"

type GuildChannel struct {
	Id       string // Discord ID
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
