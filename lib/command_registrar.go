package lib

import "github.com/bwmarrin/discordgo"

type CommandRegistry struct {
	global map[string]Command
	guild  *MapLock[map[string]Command]
}

func NewCommandRegistrar() *CommandRegistry {
	return &CommandRegistry{
		global: make(map[string]Command),
		guild:  NewMapLock[map[string]Command](),
	}
}

type Command struct {
	*discordgo.ApplicationCommand
	Respond     InteractionAction
	Authorizers []Authorizer
}

type Authorizer func(ia *InteractionArgs) error

func (cr *CommandRegistry) RegisterGlobal(c Command) {
	if _, ok := cr.global[c.Name]; ok {
		panic("Global command already registered: " + c.Name)
	}
	cr.global[c.Name] = c
}

func (cr *CommandRegistry) getGuildSet(guildId string) map[string]Command {
	guildSet, _ := cr.guild.GetOrSet(guildId, func() (map[string]Command, error) {
		return map[string]Command{}, nil
	})
	return guildSet
}

// TODO: make sure this is refreshed on server restarts
func (cr *CommandRegistry) RegisterGuild(guildId string, c Command) {
	guildSet := cr.getGuildSet(guildId)
	guildSet[c.Name] = c
}

func (cr *CommandRegistry) GetAllCommands(guildId string) map[string]Command {
	allCommands := cr.global
	guildSet := cr.getGuildSet(guildId)
	for name, cmd := range guildSet {
		allCommands[name] = cmd
	}
	return allCommands
}

func (cr *CommandRegistry) GetGlobalCommands() map[string]Command {
	return cr.global
}
