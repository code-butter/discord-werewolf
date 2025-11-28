package testlib

import (
	"discord-werewolf/lib"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type TestDiscordSessionProvider struct {
	sessions map[string]lib.DiscordSession
}

func NewTestDiscordSessionProvider(sessionMap map[string]lib.DiscordSession) *TestDiscordSessionProvider {
	return &TestDiscordSessionProvider{
		sessions: sessionMap,
	}
}

func (t TestDiscordSessionProvider) GetSession(guildId string) lib.DiscordSession {
	return t.sessions[guildId]
}

func NewTestSession(options TestSessionOptions) *TestSession {
	var id string
	if options.Id == nil {
		id = uuid.NewString()
	} else {
		id = *options.Id
	}
	return &TestSession{
		GuildId:       id,
		Name:          "Test Guild",
		GuildChannels: options.Channels,
		GuildRoles:    options.GuildRoles,
		Owner:         options.Owner,
		Members:       options.GuildMembers,
	}
}

func NewGuildTestSession(guildId string, name string, options TestSessionOptions) *TestSession {
	return &TestSession{
		GuildId:       guildId,
		Name:          name,
		GuildChannels: options.Channels,
		GuildRoles:    options.GuildRoles,
		Owner:         options.Owner,
		Members:       options.GuildMembers,
	}
}

type TestSessionOptions struct {
	Id           *string
	Channels     []*discordgo.Channel
	GuildMembers []*discordgo.Member
	GuildRoles   []*discordgo.Role
	Owner        *discordgo.User
}

type TestSession struct {
	GuildId       string
	Name          string
	GuildChannels []*discordgo.Channel
	GuildRoles    []*discordgo.Role
	Owner         *discordgo.User
	Members       []*discordgo.Member
}

func (t *TestSession) Guild() (*discordgo.Guild, error) {
	return &discordgo.Guild{
		ID:   t.GuildId,
		Name: t.Name,
	}, nil
}

func (t *TestSession) Message(_ string, _ string) error {
	return nil
}

func (t *TestSession) MessageEmbed(_ string, _ *discordgo.MessageEmbed) error {
	return nil
}

func (t *TestSession) Channels() ([]*discordgo.Channel, error) {
	return t.GuildChannels, nil
}

func (t *TestSession) CreateTextChannel(name string, parentId string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:       uuid.NewString(),
		GuildID:  t.GuildId,
		Name:     name,
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: parentId,
	}
	t.GuildChannels = append(t.GuildChannels, channel)
	return channel, nil
}

func (t *TestSession) CreateCategoryChannel(name string) (*discordgo.Channel, error) {
	channel := &discordgo.Channel{
		ID:      uuid.NewString(),
		GuildID: t.GuildId,
		Name:    name,
		Type:    discordgo.ChannelTypeGuildCategory,
	}
	t.GuildChannels = append(t.GuildChannels, channel)
	return channel, nil
}

func (t *TestSession) ClearChannelMessages(_ string) error {
	return nil
}

func (t *TestSession) GetRoles() ([]*discordgo.Role, error) {
	return t.GuildRoles, nil
}

func (t *TestSession) GetRoleByName(name string) (*discordgo.Role, error) {
	for _, role := range t.GuildRoles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, errors.New("role not found")
}

func (t *TestSession) EnsureRoleCreated(name string, color int, _ discordgo.Roles) error {
	for _, role := range t.GuildRoles {
		if role.Name == name {
			role.Color = color
			return nil
		}
	}
	t.GuildRoles = append(t.GuildRoles, &discordgo.Role{
		Name:  name,
		Color: color,
	})
	return nil
}

func (t *TestSession) DeleteChannel(id string) error {
	idx := -1
	for i, channel := range t.GuildChannels {
		if channel.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return errors.New("channel not found")
	}
	t.GuildChannels = append(t.GuildChannels[:idx], t.GuildChannels[idx+1:]...)
	return nil
}

func (t *TestSession) getMember(userId string) (*discordgo.Member, error) {
	var selectedMember *discordgo.Member
	for _, member := range t.Members {
		if member.User.ID == userId {
			selectedMember = member
			break
		}
	}
	if selectedMember == nil {
		return nil, errors.New("User with ID " + userId + " not found")
	}
	return selectedMember, nil
}

func (t *TestSession) AssignRole(userId string, roleName string) error {
	var err error
	role, err := t.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	selectedMember, err := t.getMember(userId)
	if err != nil {
		return err
	}
	selectedMember.Roles = append(selectedMember.Roles, role.ID)
	return nil
}

func (t *TestSession) RemoveRole(userId string, roleName string) error {
	var err error
	role, err := t.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	selectedMember, err := t.getMember(userId)
	if err != nil {
		return err
	}
	for idx, roleId := range selectedMember.Roles {
		if role.ID == roleId {
			selectedMember.Roles = slices.Delete(selectedMember.Roles, idx, idx+1)
		}
	}
	return nil
}

func (t *TestSession) GuildMember(id string) (*discordgo.Member, error) {
	for _, member := range t.Members {
		if member.User.ID == id {
			return member, nil
		}
	}
	return nil, errors.New("member not found")
}

func (t *TestSession) GuildMembers() ([]*discordgo.Member, error) {
	return t.Members, nil
}

func (t *TestSession) GuildMembersWithRole(roleName string) ([]*discordgo.Member, error) {
	var role, err = t.GetRoleByName(roleName)
	if err != nil {
		return nil, err
	}
	var members = make([]*discordgo.Member, 0)
	for _, guildMember := range t.Members {
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
