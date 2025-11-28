package lib

import (
	"github.com/bwmarrin/discordgo"
)

type InteractionArgs struct {
	*SessionArgs
	Interaction Interaction
	GuildId     string
}

type InteractionAction func(*InteractionArgs) error

type Interaction interface {
	// DeferredResponse Call this when potentially taking a long time to respond
	DeferredResponse(msg string, ephemeral bool) error

	// FollowupMessage Call this after doing potentially long operation
	FollowupMessage(message string, ephemeral bool) error

	// Respond Call this when sending a quick response
	Respond(message string, ephemeral bool) error

	// GuildId returns the current interaction's guild ID
	GuildId() string

	AssignRoleToRequester(roleName string) error

	RemoveRoleFromRequester(roleName string) error

	RequesterHasRole(roleName string) (bool, error)
	Requester() *discordgo.User
	CommandData() discordgo.ApplicationCommandInteractionData
	ChannelId() string
}

// TODO: make tests for live interaction with real discord server

func NewLiveInteraction(interaction *discordgo.InteractionCreate, session DiscordSession) Interaction {
	return LiveInteraction{
		session:     session,
		interaction: interaction,
	}
}

type LiveInteraction struct {
	session     DiscordSession
	interaction *discordgo.InteractionCreate
}

func (l LiveInteraction) ChannelId() string {
	return l.interaction.ChannelID
}

func (l LiveInteraction) CommandData() discordgo.ApplicationCommandInteractionData {
	return l.interaction.ApplicationCommandData()
}

func (l LiveInteraction) Requester() *discordgo.User {
	return l.interaction.Member.User
}

func (l LiveInteraction) RequesterHasRole(roleName string) (bool, error) {
	role, err := l.session.GetRoleByName(roleName)
	if err != nil {
		return false, err
	}
	for _, roleId := range l.interaction.Member.Roles {
		if roleId == role.ID {
			return true, nil
		}
	}
	return false, nil
}

func (l LiveInteraction) AssignRoleToRequester(roleName string) error {
	return l.session.AssignRole(l.interaction.Member.User.ID, roleName)
}

func (l LiveInteraction) RemoveRoleFromRequester(roleName string) error {
	return l.session.RemoveRole(l.interaction.Member.User.ID, roleName)
}

func (l LiveInteraction) DeferredResponse(msg string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return l.session.InteractionRespond(l.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   flags,
		},
	})
}

func (l LiveInteraction) FollowupMessage(message string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return l.session.FollowupMessage(l.interaction.Interaction, &discordgo.WebhookParams{
		Content: message,
		Flags:   flags,
	})
}

func (l LiveInteraction) Respond(message string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return l.session.InteractionRespond(l.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   flags,
		},
	})
}

func (l LiveInteraction) GuildId() string {
	return l.interaction.GuildID
}
