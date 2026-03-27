package lib

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type InteractionArgs struct {
	*SessionArgs
	Interaction Interaction
	GuildId     string
}

type InteractionAction func(*InteractionArgs) error

type Interaction interface {
	// SilentDeferred Call this when acknowledging an action message without an immediate response
	SilentDeferred() error

	// FollowupMessageEdit updates the action message. Automatically updates the message and channel IDs
	FollowupMessageEdit(edit discordgo.MessageEdit) error

	// CustomIdData returns the secondary custom ID for the action
	CustomIdData() string

	Message() *discordgo.Message

	// DeferredResponse Call this when potentially taking a long time to respond
	DeferredResponse(msg string, ephemeral bool) error

	// FollowupMessage call this after doing a potentially long operation
	FollowupMessage(message string, ephemeral bool) error

	// Respond Call this when sending a quick response
	Respond(message string, ephemeral bool) error

	// InteractionRespond Call this to have more control over the response
	InteractionRespond(*discordgo.InteractionResponse) error

	// GuildId returns the current interaction's guild ID
	GuildId() string

	// AssignRoleToRequester assigns a role to the requesting user
	AssignRoleToRequester(roleName string) error

	// RemoveRoleFromRequester Removes a role from the requesting user
	RemoveRoleFromRequester(roleName string) error

	//RequesterHasRole returns true if the requesting user has the given role
	RequesterHasRole(roleName string) (bool, error)
	// Requester returns the requesting user
	Requester() *discordgo.User
	// CommandData returns the data for the command
	CommandData() discordgo.ApplicationCommandInteractionData
	// MessageComponentData returns the data for the message component
	MessageComponentData() discordgo.MessageComponentInteractionData
	// ChannelId returns the current interaction's channel ID
	ChannelId() string
	ModelSubmitData() discordgo.ModalSubmitInteractionData
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

func (l LiveInteraction) ModelSubmitData() discordgo.ModalSubmitInteractionData {
	return l.interaction.ModalSubmitData()
}

func (l LiveInteraction) CustomIdData() string {
	var id string
	if l.interaction.Interaction.Type == discordgo.InteractionMessageComponent {
		id = l.interaction.MessageComponentData().CustomID
	} else if l.interaction.Interaction.Type == discordgo.InteractionModalSubmit {
		id = l.interaction.ModalSubmitData().CustomID
	} else {
		return ""
	}
	parts := strings.SplitN(id, ":", 2)
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (l LiveInteraction) FollowupMessageEdit(edit discordgo.MessageEdit) error {
	if l.interaction.Interaction.Type != discordgo.InteractionMessageComponent {
		return errors.New("cannot edit followup message for non-message component interaction")
	}
	msg := l.interaction.Message
	if msg == nil {
		return errors.New("there is no message to edit")
	}
	edit.ID = msg.ID
	edit.Channel = msg.ChannelID
	return l.session.MessageEditComplex(&edit)
}

func (l LiveInteraction) Message() *discordgo.Message {
	return l.interaction.Message
}

func (l LiveInteraction) InteractionRespond(ir *discordgo.InteractionResponse) error {
	return l.session.InteractionRespond(l.interaction.Interaction, ir)
}

func (l LiveInteraction) SilentDeferred() error {
	return l.session.InteractionRespond(l.interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
}

func (l LiveInteraction) MessageComponentData() discordgo.MessageComponentInteractionData {
	return l.interaction.MessageComponentData()
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
