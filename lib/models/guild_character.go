package models

import (
	"database/sql/driver"
	"discord-werewolf/lib"
	"encoding/json"
)

const GameStatusPlaying = 1
const GameStatusAlive = 2
const GameStatusDead = 3

const EffectInjured = 1
const EffectStunned = 2
const EffectMuted = 4
const EffectInLove = 8

const SecondaryVampire = 1
const SecondaryHenchman = 2

const CharacterVillager = 1
const CharacterWolf = 2
const CharacterWolfCub = 3
const CharacterWitch = 4
const CharacterSeer = 5
const CharacterFool = 6
const CharacterSeerApprentice = 7
const CharacterBodyguard = 8
const CharacterLycan = 9
const CharacterDoppelganger = 10
const CharacterChaosDemon = 11
const CharacterVampireKing = 12

type GuildCharacter struct {
	Id                   int `gorm:"primaryKey"`
	GuildId              int `gorm:"primaryKey"`
	CharacterId          int
	SecondaryCharacterId int
	GameStatus           int
	EffectMask           int
	ExtraData            CharacterExtraData
}

type CharacterExtraData map[string]interface{}

func (m CharacterExtraData) Value() (driver.Value, error) {
	return json.Marshal(m)
}
func (m *CharacterExtraData) Scan(value interface{}) error {
	return lib.UnMarshalBytes(m, value)
}

func (m CharacterExtraData) GormDataType() string {
	return "character_extra_data"
}
