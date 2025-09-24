package lib

import "github.com/bwmarrin/discordgo"

var commands = map[string]Command{}

type Command struct {
	*discordgo.ApplicationCommand
	Global  bool // Does not have any server-specific parameters
	Respond InteractionAction
}

func RegisterCommand(c Command) {
	commands[c.Name] = c
}

func GetCommands() map[string]Command {
	return commands
}
