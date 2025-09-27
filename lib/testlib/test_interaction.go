package testlib

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type TestInteraction struct {
	guildId  string
	name     string
	channels []*discordgo.Channel
	roles    []*discordgo.Role
}

func (d *TestInteraction) AssignRole(userId string, roleName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *TestInteraction) AssignRoleToRequester(roleName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *TestInteraction) RemoveRole(userId string, roleName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *TestInteraction) RemoveRoleFromRequester(roleName string) error {
	//TODO implement me
	panic("implement me")
}

func (d *TestInteraction) GetRoles() (discordgo.Roles, error) {
	return d.roles, nil
}

func (d *TestInteraction) EnsureRoleCreated(name string, color int, _ discordgo.Roles) error {
	for _, role := range d.roles {
		if role.Name == name {
			role.Color = color
			return nil
		}
	}
	d.roles = append(d.roles, &discordgo.Role{
		Name:  name,
		Color: color,
	})
	return nil
}

func (d *TestInteraction) DeleteChannel(id string) error {
	idx := -1
	for i, channel := range d.channels {
		if channel.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("channel not found")
	}
	d.channels = append(d.channels[:idx], d.channels[idx+1:]...)
	return nil
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
