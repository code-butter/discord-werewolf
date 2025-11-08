package testlib

import (
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func NewTestSession(guildId string, name string, options TestSessionOptions) *TestSession {
	return &TestSession{
		guildId:      guildId,
		name:         name,
		channels:     options.Channels,
		guildRoles:   options.GuildRoles,
		userRoles:    options.UserRoles,
		owner:        options.Owner,
		guildMembers: options.GuildMembers,
	}
}

type TestSessionOptions struct {
	Channels     []*discordgo.Channel
	GuildMembers []*discordgo.Member
	GuildRoles   []*discordgo.Role
	UserRoles    map[string][]string
	Owner        *discordgo.User
}

type TestSession struct {
	guildId      string
	name         string
	channels     []*discordgo.Channel
	guildRoles   []*discordgo.Role
	userRoles    map[string][]string
	owner        *discordgo.User
	guildMembers []*discordgo.Member
}

func (t *TestSession) Guild() (*discordgo.Guild, error) {
	return &discordgo.Guild{
		ID:   t.guildId,
		Name: t.name,
	}, nil
}

func (t *TestSession) Message(_ string, _ string) error {
	return nil
}

func (t *TestSession) MessageEmbed(_ string, _ *discordgo.MessageEmbed) error {
	return nil
}

func (t *TestSession) Channels() ([]*discordgo.Channel, error) {
	return t.channels, nil
}

func (t *TestSession) CreateTextChannel(name string, parentId string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:       uuid.NewString(),
		GuildID:  t.guildId,
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: parentId,
	}
	t.channels = append(t.channels, channel)
	return channel, nil
}

func (t *TestSession) CreateCategoryChannel(name string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:      uuid.NewString(),
		GuildID: t.guildId,
		Name:    name,
		Type:    discordgo.ChannelTypeGuildCategory,
	}
	t.channels = append(t.channels, channel)
	return channel, nil
}

func (t *TestSession) ClearChannelMessages(_ string) error {
	return nil
}

func (t *TestSession) GetRoles() ([]*discordgo.Role, error) {
	return t.guildRoles, nil
}

func (t *TestSession) GetRoleByName(name string) (*discordgo.Role, error) {
	for _, role := range t.guildRoles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, errors.New("role not found")
}

func (t *TestSession) EnsureRoleCreated(name string, color int, _ discordgo.Roles) error {
	for _, role := range t.guildRoles {
		if role.Name == name {
			role.Color = color
			return nil
		}
	}
	t.guildRoles = append(t.guildRoles, &discordgo.Role{
		Name:  name,
		Color: color,
	})
	return nil
}

func (t *TestSession) DeleteChannel(id string) error {
	idx := -1
	for i, channel := range t.channels {
		if channel.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("channel not found")
	}
	t.channels = append(t.channels[:idx], t.channels[idx+1:]...)
	return nil
}

func (t *TestSession) AssignRole(userId string, roleName string) error {
	role, err := t.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	t.userRoles[userId] = append(t.userRoles[userId], role.ID)
	return nil
}

func (t *TestSession) RemoveRole(userId string, roleName string) error {
	role, err := t.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	for idx, roleId := range t.userRoles[userId] {
		if role.ID == roleId {
			t.userRoles[userId] = append(t.userRoles[userId][:idx], t.userRoles[userId][idx+1:]...)
		}
	}
	return nil
}

func (t *TestSession) GuildMembers() ([]*discordgo.Member, error) {
	return t.guildMembers, nil
}

func (t *TestSession) GuildMembersWithRole(roleName string) ([]*discordgo.Member, error) {
	var role, err = t.GetRoleByName(roleName)
	if err != nil {
		return nil, err
	}
	var members = make([]*discordgo.Member, 0)
	for _, guildMember := range t.guildMembers {
		for _, mRoleId := range guildMember.Roles {
			if mRoleId == role.ID {
				members = append(members, guildMember)
			}
		}
	}
	return members, nil
}

func (t *TestSession) InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error {
	return nil
}

func (t *TestSession) FollowupMessage(interaction *discordgo.Interaction, params *discordgo.WebhookParams) error {
	return nil
}

func (t *TestSession) RoleChannelPermissions(channelId string, roleId string, allow, deny int64) error {
	return nil
}

func (t *TestSession) UserChannelPermissions(channelId string, userId string, allow, deny int64) error {
	return nil
}

func (t *TestSession) DeleteChannelOverridePermissions(channelId string, id string) error {
	return nil
}
