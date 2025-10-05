package models

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
	Id                   string `gorm:"primaryKey"`
	GuildId              string `gorm:"primaryKey"`
	CharacterId          int
	SecondaryCharacterId int
	EffectMask           int
	ExtraData            JsonMap
}
