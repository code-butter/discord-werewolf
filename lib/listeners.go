package lib

import (
	"discord-werewolf/lib/models"

	"github.com/bwmarrin/discordgo"
)

func NewGameListeners() *GameListeners {
	return &GameListeners{
		GameStart:      &ListenerTracker[GameStartData]{},
		NightStart:     &ListenerTracker[NightStartData]{},
		DayStart:       &ListenerTracker[DayStartData]{},
		CharacterDeath: &ListenerTracker[CharacterDeathData]{},
		GameEnd:        &ListenerTracker[GameEndData]{},
	}
}

type ListenerCallback[T any] func(s *SessionArgs, data T) error

type ListenerTracker[T any] struct {
	listeners []ListenerCallback[T]
}

type GameListeners struct {
	GameStart      *ListenerTracker[GameStartData]
	NightStart     *ListenerTracker[NightStartData]
	DayStart       *ListenerTracker[DayStartData]
	CharacterDeath *ListenerTracker[CharacterDeathData]
	GameEnd        *ListenerTracker[GameEndData]
}

func (lt *ListenerTracker[T]) Add(lc ListenerCallback[T]) {
	lt.listeners = append(lt.listeners, lc)
}

func (lt *ListenerTracker[T]) Trigger(s *SessionArgs, data T) error {
	for _, listener := range lt.listeners {
		if err := listener(s, data); err != nil {
			return err
		}
	}
	return nil
}

type GameStartData struct {
	Guild      *models.Guild
	Characters []models.GuildCharacter
}
type NightStartData struct {
	Guild *models.Guild
}
type DayStartData struct {
	Guild *models.Guild
}
type CharacterDeathData struct {
	Guild  *models.Guild
	Target models.GuildCharacter
}

type GameEndData struct {
	Guild *models.Guild
}

type GameOver struct {
	MessageEmbed discordgo.MessageEmbed
}

func (g GameOver) Error() string {
	return "Game over, dude." // If you see this error anywhere, it's a bug. Return the embedded message to the server instead.
}
