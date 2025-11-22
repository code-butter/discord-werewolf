package testlib

import (
	"discord-werewolf/lib"

	"github.com/bwmarrin/discordgo"
)

func NewTestInteraction(args lib.SessionArgs, options TestInteractionOptions) *TestInteraction {
	return &TestInteraction{
		session:     args.Session,
		requester:   options.Requester,
		commandData: options.CommandData,
		channelId:   options.ChannelId,
	}
}

type TestInteractionOptions struct {
	Requester   *discordgo.User
	CommandData discordgo.ApplicationCommandInteractionData
	ChannelId   string
}

type TestInteraction struct {
	session     lib.DiscordSession
	requester   *discordgo.User
	commandData discordgo.ApplicationCommandInteractionData
	channelId   string
}

func (d *TestInteraction) ChannelId() string {
	return d.channelId
}

func (d *TestInteraction) GuildId() string {
	guild, err := d.session.Guild()
	if err != nil {
		panic(err)
	}
	return guild.ID
}

func (d *TestInteraction) AssignRoleToRequester(roleName string) error {
	return d.session.AssignRole(d.requester.ID, roleName)
}

func (d *TestInteraction) RemoveRoleFromRequester(roleName string) error {
	return d.session.RemoveRole(d.requester.ID, roleName)
}

func (d *TestInteraction) DeferredResponse(string, bool) error {
	return nil
}

func (d *TestInteraction) CommandData() discordgo.ApplicationCommandInteractionData {
	return d.commandData
}

func (d *TestInteraction) Requester() *discordgo.User {
	return d.requester
}

func (d *TestInteraction) RequesterHasRole(roleName string) (bool, error) {
	membersWithRole, err := d.session.GuildMembersWithRole(roleName)
	if err != nil {
		return false, err
	}
	for _, member := range membersWithRole {
		if member.User.ID == d.requester.ID {
			return true, nil
		}
	}
	return false, nil
}

func (d *TestInteraction) FollowupMessage(string, bool) error {
	return nil
}

func (d *TestInteraction) Respond(string, bool) error {
	return nil
}
