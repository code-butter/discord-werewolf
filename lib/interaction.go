package lib

import "github.com/bwmarrin/discordgo"

type InteractionAction func(i Interaction) error

type Interaction interface {
	// DeferredResponse Call this when potentially taking a long time to respond
	DeferredResponse() error

	// FollowupMessage Call this after doing potentially long operation
	FollowupMessage(message string, ephemeral bool) error

	// Respond Call this when sending a quick response
	Respond(message string, ephemeral bool) error

	// GuildId returns the current interaction's guild ID
	GuildId() string

	// Guild gets the current interaction's guild object
	Guild() (*discordgo.Guild, error)

	// Channels gets all channels from the current guild
	Channels() ([]*discordgo.Channel, error)

	// CreateTextChannel Creates a text channel, optionally within a category
	CreateTextChannel(name string, parentId string) (*discordgo.Channel, error)

	// CreateCategoryChannel Creates a channel category
	CreateCategoryChannel(name string) (*discordgo.Channel, error)
}

type LiveInteraction struct {
	Session           *discordgo.Session
	InteractionCreate *discordgo.InteractionCreate
}

func (l LiveInteraction) DeferredResponse() error {
	return l.Session.InteractionRespond(l.InteractionCreate.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

func (l LiveInteraction) FollowupMessage(message string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	_, err := l.Session.FollowupMessageCreate(l.InteractionCreate.Interaction, false, &discordgo.WebhookParams{
		Content: message,
		Flags:   flags,
	})
	return err
}

func (l LiveInteraction) Respond(message string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return l.Session.InteractionRespond(l.InteractionCreate.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
			Flags:   flags,
		},
	})
}

func (l LiveInteraction) GuildId() string {
	return l.InteractionCreate.GuildID
}

func (l LiveInteraction) Guild() (*discordgo.Guild, error) {
	guild, err := l.Session.State.Guild(l.InteractionCreate.GuildID)
	if err != nil {
		guild, err = l.Session.Guild(l.InteractionCreate.GuildID)
		if err != nil {
			return nil, err
		}
	}
	return guild, nil
}

func (l LiveInteraction) Channels() ([]*discordgo.Channel, error) {
	return l.Session.GuildChannels(l.InteractionCreate.GuildID)
}

func (l LiveInteraction) CreateTextChannel(name string, parentId string) (*discordgo.Channel, error) {
	return l.Session.GuildChannelCreateComplex(l.InteractionCreate.GuildID, discordgo.GuildChannelCreateData{
		Name:     name,
		ParentID: parentId,
		Type:     discordgo.ChannelTypeGuildText,
	})
}

func (l LiveInteraction) CreateCategoryChannel(name string) (*discordgo.Channel, error) {
	return l.Session.GuildChannelCreate(l.InteractionCreate.GuildID, name, discordgo.ChannelTypeGuildCategory)
}
