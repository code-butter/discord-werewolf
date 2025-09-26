package testlib

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type TestInteraction struct {
	guildId  string
	name     string
	channels []*discordgo.Channel
}

func NewTestInteraction(guildId string, name string, channels []*discordgo.Channel) *TestInteraction {
	return &TestInteraction{
		guildId:  guildId,
		name:     name,
		channels: channels,
	}
}

func (d *TestInteraction) DeferredResponse() error {
	return nil
}

func (d *TestInteraction) FollowupMessage(string, bool) error {
	return nil
}

func (d *TestInteraction) Respond(string, bool) error {
	return nil
}

func (d *TestInteraction) GuildId() string {
	return d.guildId
}

func (d *TestInteraction) Guild() (*discordgo.Guild, error) {
	return &discordgo.Guild{
		ID:   d.guildId,
		Name: d.name,
	}, nil
}

func (d *TestInteraction) Channels() ([]*discordgo.Channel, error) {
	return d.channels, nil
}

func (d *TestInteraction) CreateTextChannel(name string, parentId string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:       uuid.NewString(),
		GuildID:  d.guildId,
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: parentId,
	}
	d.channels = append(d.channels, channel)
	return channel, nil
}

func (d *TestInteraction) CreateCategoryChannel(name string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:      uuid.NewString(),
		GuildID: d.guildId,
		Name:    name,
		Type:    discordgo.ChannelTypeGuildCategory,
	}
	d.channels = append(d.channels, channel)
	return channel, nil
}
