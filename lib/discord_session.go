package lib

import (
	"discord-werewolf/lib/models"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

// TODO: implement retry-after logic for API limits and service outages
// TODO: add context listener for application shutdown

type SessionArgs struct {
	Session  DiscordSession
	Injector *do.Injector
	guild    *models.Guild
}

func (args *SessionArgs) AppGuild() (*models.Guild, error) {
	if args.guild == nil {
		db := do.MustInvoke[*gorm.DB](args.Injector)
		guild, err := args.Session.Guild()
		if err != nil {
			return nil, err
		}
		result := db.Model((*models.Guild)(nil)).Where("id = ?", guild.ID).First(&args.guild)
		if result.Error != nil {
			return nil, result.Error
		}
	}
	return args.guild, nil
}

func (args *SessionArgs) GuildCharacters() ([]*models.GuildCharacter, error) {
	gormDB := do.MustInvoke[*gorm.DB](args.Injector)
	guild, err := args.Session.Guild()
	if err != nil {
		return nil, err
	}
	var characters []*models.GuildCharacter
	result := gormDB.
		Where("guild_id = ?", guild.ID).
		Find(&characters)
	if result.Error != nil {
		return nil, result.Error
	}
	return characters, nil
}

func (args *SessionArgs) GuildCharacter(id string) (*models.GuildCharacter, error) {
	gormDB := do.MustInvoke[*gorm.DB](args.Injector)
	guild, err := args.Session.Guild()
	if err != nil {
		return nil, err
	}
	var character models.GuildCharacter
	result := gormDB.
		Where("guild_id = ? AND id = ?", guild.ID, id).
		First(&character)
	if result.Error != nil {
		return nil, result.Error
	}
	return &character, nil
}

type DiscordSession interface {
	// Guild gets the server's guild object
	Guild() (*discordgo.Guild, error)

	Message(channelId string, message string) error
	MessageEmbed(channelId string, embed *discordgo.MessageEmbed) error

	// Channels gets all channels from the current guild
	Channels() ([]*discordgo.Channel, error)

	// CreateTextChannel Creates a text channel, optionally within a category
	CreateTextChannel(name string, parentId string) (*discordgo.Channel, error)

	// CreateCategoryChannel Creates a channel category
	CreateCategoryChannel(name string) (*discordgo.Channel, error)

	// ClearChannelMessages removes all messages from a channel
	ClearChannelMessages(id string) error

	// GetRoles get all current roles for the guild.
	GetRoles() ([]*discordgo.Role, error)

	// GetRoleByName gets a role by its name
	GetRoleByName(name string) (*discordgo.Role, error)

	// EnsureRoleCreated Created or updates role with color. Pass in roles from `GetRoles`.
	EnsureRoleCreated(name string, color int, roles discordgo.Roles) error

	// DeleteChannel Removes discord channel
	DeleteChannel(id string) error

	// AssignRole gives user a role
	AssignRole(userId string, roleName string) error

	// RemoveRole removes role from user
	RemoveRole(userId string, roleName string) error
	GuildMember(id string) (*discordgo.Member, error)
	GuildMembers() ([]*discordgo.Member, error)
	GuildMembersWithRole(roleName string) ([]*discordgo.Member, error)

	InteractionRespond(*discordgo.Interaction, *discordgo.InteractionResponse) error
	FollowupMessage(*discordgo.Interaction, *discordgo.WebhookParams) error

	RoleChannelPermissions(channelId string, roleId string, allow, deny int64) error
	UserChannelPermissions(channelId string, userId string, allow, deny int64) error
	DeleteChannelOverridePermissions(channelId string, id string) error
}

type DiscordSessionProvider interface {
	GetSession(guildId string) DiscordSession
}

type LiveDiscordSessionProvider struct {
	getter  GuildSessionGetter
	mapLock *MapLock[DiscordSession]
}

type GuildSessionGetter func(guildId string) (DiscordSession, error)

func NewDiscordSessionProvider(getter GuildSessionGetter) *LiveDiscordSessionProvider {
	return &LiveDiscordSessionProvider{
		getter:  getter,
		mapLock: NewMapLock[DiscordSession](),
	}
}

func (g *LiveDiscordSessionProvider) GetSession(guildId string) DiscordSession {
	value, _ := g.mapLock.GetOrSet(guildId, func() (DiscordSession, error) {
		return g.getter(guildId)
	})
	return value
}

func NewGuildDiscordSession(guildId string, session *discordgo.Session, cacheTimeout time.Duration, clock Clock) *GuildDiscordSession {
	return &GuildDiscordSession{
		guildID:   guildId,
		session:   session,
		roleCache: NewInteractionCache[[]*discordgo.Role](cacheTimeout, clock),
	}
}

type GuildDiscordSession struct {
	session   *discordgo.Session
	guildID   string
	roleCache *InteractionCache[[]*discordgo.Role]
}

func (l *GuildDiscordSession) GuildMember(userId string) (*discordgo.Member, error) {
	return l.session.GuildMember(l.guildID, userId)
}

func (l *GuildDiscordSession) DeleteChannelOverridePermissions(channelId string, id string) error {
	return l.session.ChannelPermissionDelete(channelId, id)
}

func (l *GuildDiscordSession) RoleChannelPermissions(channelId string, roleId string, allow, deny int64) error {
	return l.session.ChannelPermissionSet(
		channelId,
		roleId,
		discordgo.PermissionOverwriteTypeRole,
		allow,
		deny,
	)
}

func (l *GuildDiscordSession) UserChannelPermissions(channelId string, roleId string, allow, deny int64) error {
	return l.session.ChannelPermissionSet(
		channelId,
		roleId,
		discordgo.PermissionOverwriteTypeMember,
		allow,
		deny,
	)
}

func (l *GuildDiscordSession) ClearChannelMessages(channelId string) error {
	const maxMessages = 100
	const fourteenDays = 14 * 24 * time.Hour
	var before string
	for {
		messages, err := l.session.ChannelMessages(channelId, maxMessages, before, "", "")
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}
		var bulkDelete []string
		for _, message := range messages {
			if time.Since(message.Timestamp) < fourteenDays {
				bulkDelete = append(bulkDelete, message.ID)
			} else {
				if err = l.session.ChannelMessageDelete(message.ChannelID, message.ID); err != nil {
					return err
				}
			}
		}
		toDeleteCount := len(bulkDelete)
		if toDeleteCount > 1 {
			if err = l.session.ChannelMessagesBulkDelete(channelId, bulkDelete); err != nil {
				return err
			}
		} else if toDeleteCount == 1 {
			if err = l.session.ChannelMessageDelete(channelId, bulkDelete[0]); err != nil {
				return err
			}
		}
		before = messages[len(messages)-1].ID
	}
}

func (l *GuildDiscordSession) InteractionRespond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error {
	return l.session.InteractionRespond(interaction, response)
}

func (l *GuildDiscordSession) FollowupMessage(interaction *discordgo.Interaction, params *discordgo.WebhookParams) error {
	_, err := l.session.FollowupMessageCreate(interaction, false, params)
	return err
}

func (l *GuildDiscordSession) Message(channelId string, message string) error {
	_, err := l.session.ChannelMessageSend(channelId, message)
	return err
}

func (l *GuildDiscordSession) MessageEmbed(channelId string, embed *discordgo.MessageEmbed) error {
	_, err := l.session.ChannelMessageSendEmbed(channelId, embed)
	return err
}

func (l *GuildDiscordSession) GuildMembers() ([]*discordgo.Member, error) {
	var members []*discordgo.Member
	afterId := ""
	for {
		batch, err := l.session.GuildMembers(l.guildID, afterId, 1000)
		if err != nil {
			return nil, err
		}
		batchLen := len(batch)
		members = append(members, batch...)
		if batchLen < 1000 {
			break
		}
		afterId = batch[batchLen-1].User.ID
	}
	return members, nil
}

func (l *GuildDiscordSession) GetRoles() ([]*discordgo.Role, error) {
	if roles, ok := l.roleCache.Get(l.guildID); ok {
		return *roles, nil
	}
	return l.session.GuildRoles(l.guildID)
}

func (l *GuildDiscordSession) GetRoleByName(name string) (*discordgo.Role, error) {
	roles, err := l.GetRoles()
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		if role.Name == name {
			return role, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("role not found: %s", name))
}

func (l *GuildDiscordSession) AssignRole(userId string, roleName string) error {
	role, err := l.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	return l.session.GuildMemberRoleAdd(l.guildID, userId, role.ID)
}

func (l *GuildDiscordSession) GuildMembersWithRole(roleName string) ([]*discordgo.Member, error) {
	members, err := l.GuildMembers()
	if err != nil {
		return nil, err
	}
	role, err := l.GetRoleByName(roleName)
	if err != nil {
		return nil, err
	}
	roleMembers := make([]*discordgo.Member, 0)
	for _, member := range members {
		for _, r := range member.Roles {
			if r == role.ID {
				roleMembers = append(roleMembers, member)
				break
			}
		}
	}
	return roleMembers, nil
}

func (l *GuildDiscordSession) RemoveRole(userId string, roleName string) error {
	role, err := l.GetRoleByName(roleName)
	if err != nil {
		return err
	}
	return l.session.GuildMemberRoleRemove(l.guildID, userId, role.ID)
}

func (l *GuildDiscordSession) Guild() (*discordgo.Guild, error) {
	guild, err := l.session.State.Guild(l.guildID)
	if err != nil {
		guild, err = l.session.Guild(l.guildID)
		if err != nil {
			return nil, err
		}
	}
	return guild, nil
}

func (l *GuildDiscordSession) Channels() ([]*discordgo.Channel, error) {
	return l.session.GuildChannels(l.guildID)
}

func (l *GuildDiscordSession) CreateTextChannel(name string, parentId string) (*discordgo.Channel, error) {
	return l.session.GuildChannelCreateComplex(l.guildID, discordgo.GuildChannelCreateData{
		Name:     name,
		ParentID: parentId,
		Type:     discordgo.ChannelTypeGuildText,
	})
}

func (l *GuildDiscordSession) CreateCategoryChannel(name string) (*discordgo.Channel, error) {
	return l.session.GuildChannelCreate(l.guildID, name, discordgo.ChannelTypeGuildCategory)
}

func (l *GuildDiscordSession) DeleteChannel(id string) error {
	_, err := l.session.ChannelDelete(id)
	return err
}

func (l *GuildDiscordSession) EnsureRoleCreated(name string, color int, roles discordgo.Roles) error {
	var foundRole *discordgo.Role
	for _, role := range roles {
		if name == role.Name && color == role.Color {
			return nil
		} else if name == role.Name {
			foundRole = role
			break
		}
	}
	var err error
	if foundRole != nil {
		_, err = l.session.GuildRoleEdit(l.guildID, foundRole.ID, &discordgo.RoleParams{
			Color: &color,
		})
	} else {
		_, err = l.session.GuildRoleCreate(l.guildID, &discordgo.RoleParams{
			Name:  name,
			Color: &color,
		})
	}
	return err
}
