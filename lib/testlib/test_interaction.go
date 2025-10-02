package testlib

import (
	"errors"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func NewTestInteraction(guildId string, name string, options TestInteractionOptions) *TestInteraction {
	return &TestInteraction{
		guildId:    guildId,
		name:       name,
		channels:   options.Channels,
		guildRoles: options.GuildRoles,
		requester:  options.Requester,
		owner:      options.Owner,
	}
}

type TestInteractionOptions struct {
	Channels   []*discordgo.Channel
	GuildRoles []*discordgo.Role
	UserRoles  []*discordgo.Role
	Owner      *discordgo.User
	Requester  *discordgo.User
}

type TestInteraction struct {
	guildId    string
	name       string
	channels   []*discordgo.Channel
	guildRoles []*discordgo.Role
	userRoles  []*discordgo.Role
	owner      *discordgo.User
	requester  *discordgo.User
}

func (d *TestInteraction) CommandData() discordgo.ApplicationCommandInteractionData {
	//TODO implement me
	panic("implement me")
}

func (d *TestInteraction) Requester() *discordgo.User {
	return d.Requester()
}
func (d *TestInteraction) GetRoles() ([]*discordgo.Role, error) {
	return d.guildRoles, nil
}

func (d *TestInteraction) RequesterHasRole(roleName string) (bool, error) {
	for _, role := range d.userRoles {
		if role.Name == roleName {
			return true, nil
		}
	}
	return false, nil
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

func (d *TestInteraction) EnsureRoleCreated(name string, color int, _ discordgo.Roles) error {
	for _, role := range d.guildRoles {
		if role.Name == name {
			role.Color = color
			return nil
		}
	}
	d.guildRoles = append(d.guildRoles, &discordgo.Role{
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
