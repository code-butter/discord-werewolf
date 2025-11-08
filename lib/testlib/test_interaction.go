package testlib

import (
	"discord-werewolf/lib"

	"github.com/bwmarrin/discordgo"
)

func NewTestInteraction(session lib.DiscordSession, options TestInteractionOptions) *TestInteraction {
	return &TestInteraction{
		session:     session,
		requester:   options.Requester,
		commandData: options.CommandData,
	}
}

type TestInteractionOptions struct {
	Requester   *discordgo.User
	UserRoles   []*discordgo.Role
	CommandData discordgo.ApplicationCommandInteractionData
}

type TestInteraction struct {
	session     lib.DiscordSession
	userRoles   []*discordgo.Role
	requester   *discordgo.User
	commandData discordgo.ApplicationCommandInteractionData
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
	for _, role := range d.userRoles {
		if role.Name == roleName {
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
