package lib

import "github.com/bwmarrin/discordgo"

type CommandRegistrar struct {
	global map[string]Command
	guild  *MapLock[map[string]Command]
}

func NewCommandRegistrar() *CommandRegistrar {
	return &CommandRegistrar{
		global: make(map[string]Command),
		guild:  NewMapLock[map[string]Command](),
	}
}

type Command struct {
	*discordgo.ApplicationCommand
	Respond     InteractionAction
	Authorizers []CommandAuthorizer
}

type CommandAuthorizer func(ia *InteractionArgs) error

func (cr *CommandRegistrar) RegisterGlobal(c Command) {
	if _, ok := cr.global[c.Name]; ok {
		panic("Global command already registered: " + c.Name)
	}
	cr.global[c.Name] = c
}

func (cr *CommandRegistrar) getGuildSet(guildId string) map[string]Command {
	guildSet, _ := cr.guild.GetOrSet(guildId, func() (map[string]Command, error) {
		return map[string]Command{}, nil
	})
	return guildSet
}

func (cr *CommandRegistrar) RegisterGuild(guildId string, c Command) {
	guildSet := cr.getGuildSet(guildId)
	guildSet[c.Name] = c
}

// TODO: cache the results here so we're not looping in hot paths
func (cr *CommandRegistrar) GetAllCommands(guildId string) map[string]Command {
	allCommands := cr.global
	guildSet := cr.getGuildSet(guildId)
	for name, cmd := range guildSet {
		allCommands[name] = cmd
	}
	return allCommands
}

func (cr *CommandRegistrar) GetGlobalCommands() map[string]Command {
	return cr.global
}
