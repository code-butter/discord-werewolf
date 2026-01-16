package models

import (
	"discord-werewolf/lib/characters"
	"fmt"
)

type GuildCharacter struct {
	Id                   string `gorm:"primaryKey"` // The discord user ID
	GuildId              string `gorm:"primaryKey"`
	CharacterId          string
	SecondaryCharacterId string
	FakeCharacterId      string
	EffectMask           int
	ExtraData            JsonMap
}

func (gc *GuildCharacter) Label(reported bool, full bool) string {
	if reported && gc.FakeCharacterId != "" {
		return characters.Characters[gc.FakeCharacterId].Label
	}
	character := characters.Characters[gc.CharacterId]
	desc := character.Label
	if gc.SecondaryCharacterId != "" && full {
		secondaryCharacter := characters.Characters[gc.SecondaryCharacterId]
		desc += fmt.Sprintf(" (who is a %s!)", secondaryCharacter.Label)
	}
	return desc
}
