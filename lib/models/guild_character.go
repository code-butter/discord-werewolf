package models

import "log"

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

var characterDescriptions map[int]string

func init() {
	cd := map[int]string{}
	cd[CharacterVillager] = "Villager"
	cd[CharacterWolf] = "Werewolf"
	cd[CharacterWolfCub] = "Werewolf Cub"
	characterDescriptions = cd
}

type GuildCharacter struct {
	Id                   string `gorm:"primaryKey"` // The discord user ID
	GuildId              string `gorm:"primaryKey"`
	CharacterId          int
	SecondaryCharacterId int
	EffectMask           int
	ExtraData            JsonMap
}

func (gc *GuildCharacter) CharacterDescription() string {
	d, ok := characterDescriptions[gc.CharacterId]
	if !ok {
		log.Printf("Could not find description for Character ID %d", gc.CharacterId)
		return ""
	}
	return d
}
