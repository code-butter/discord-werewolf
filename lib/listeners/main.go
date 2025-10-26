package listeners

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
)

type ListenerCallback[T any] func(s *lib.SessionArgs, data T) error

type listenerTracker[T any] struct {
	listeners []ListenerCallback[T]
}

func (lt *listenerTracker[T]) Add(lc ListenerCallback[T]) {
	lt.listeners = append(lt.listeners, lc)
}

func (lt *listenerTracker[T]) Trigger(s *lib.SessionArgs, data T) error {
	for _, listener := range lt.listeners {
		if err := listener(s, data); err != nil {
			return err
		}
	}
	return nil
}

type GameStartData struct {
	Guild      models.Guild
	Characters []models.GuildCharacter
}
type NightStartData struct {
	Guild models.Guild
}
type DayStartData struct {
	Guild models.Guild
}
type CharacterDeathData struct {
	Guild  models.Guild
	Target models.GuildCharacter
	Cause  string
}

var GameStartListeners *listenerTracker[GameStartData]
var NightStartListeners *listenerTracker[NightStartData]
var DayStartListeners *listenerTracker[DayStartData]
var CharacterDeathListeners *listenerTracker[CharacterDeathData]

func init() {
	GameStartListeners = &listenerTracker[GameStartData]{}
	NightStartListeners = &listenerTracker[NightStartData]{}
	DayStartListeners = &listenerTracker[DayStartData]{}
	CharacterDeathListeners = &listenerTracker[CharacterDeathData]{}
}
