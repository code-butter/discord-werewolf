package lib

import "github.com/bwmarrin/discordgo"

var globalCommands = map[string]Command{}

type Command struct {
	*discordgo.ApplicationCommand
	Respond     InteractionAction
	Authorizers []CommandAuthorizer
}

type CommandAuthorizer func(ia *InteractionArgs) error

func RegisterGlobalCommand(c Command) {
	if _, ok := globalCommands[c.Name]; ok {
		panic("Command " + c.Name + " already registered")
	}
	globalCommands[c.Name] = c
}

func GetGlobalCommands() map[string]Command {
	return globalCommands
}
