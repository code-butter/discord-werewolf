package game_management

import (
	"discord-werewolf/lib"

	"github.com/bwmarrin/discordgo"
)

func Setup() error {

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "playing",
			Description: "Signs up for the next round.",
		},
		Global:  true,
		Respond: playing,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "stop_playing",
			Description: "Removes yourself from playing next round.",
		},
		Global:  true,
		Respond: stopPlaying,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "start_game",
			Description: "Starts the game server.",
		},
		Respond: startGame,
	})

	return nil
}

func startGame(i lib.Interaction) error {
	panic("Implement me!")
}

func playing(i lib.Interaction) error {
	if err := i.AssignRoleToRequester(lib.RolePlaying); err != nil {
		return err
	}
	return i.Respond("Now playing!", false)
}

func stopPlaying(i lib.Interaction) error {
	if err := i.RemoveRoleFromRequester(lib.RolePlaying); err != nil {
		return err
	}
	return i.Respond("Stopped playing.", false)
}
