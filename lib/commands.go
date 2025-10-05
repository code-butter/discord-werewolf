package lib

import "github.com/bwmarrin/discordgo"

var commands = map[string]Command{}

type Command struct {
	*discordgo.ApplicationCommand
	Global      bool // Does not have any server/game specific parameters
	Respond     InteractionAction
	Authorizers []CommandAuthorizer
}

type CommandAuthorizer func(ia *InteractionArgs) (bool, error)

func RegisterCommand(c Command) {
	if _, ok := commands[c.Name]; ok {
		panic("Command " + c.Name + " already registered")
	}
	commands[c.Name] = c
}

func GetCommands() map[string]Command {
	return commands
}
