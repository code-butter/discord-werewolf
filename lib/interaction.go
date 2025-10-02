package lib

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type InteractionAction func(i Interaction) error

// TODO: split this into smaller interfaces for different aspects?

type Interaction interface {
	// DeferredResponse Call this when potentially taking a long time to respond
	DeferredResponse(msg string, ephemeral bool) error

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

	// GetRoles get all current roles for the guild.
	GetRoles() ([]*discordgo.Role, error)

	// EnsureRoleCreated Created or updates role with color. Pass in roles from `GetRoles`.
	EnsureRoleCreated(name string, color int, roles discordgo.Roles) error

	// DeleteChannel Removes discord channel
	DeleteChannel(id string) error

	// TODO: the mix of global/per-user

	// AssignRole gives user a role
	AssignRole(userId string, roleName string) error

	AssignRoleToRequester(roleName string) error

	// RemoveRole removes role from user
	RemoveRole(userId string, roleName string) error

	RemoveRoleFromRequester(roleName string) error

	RequesterHasRole(roleName string) (bool, error)
	Requester() *discordgo.User
	CommandData() discordgo.ApplicationCommandInteractionData
}

// TODO: make tests for live interaction with real discord server

type LiveInteraction struct {
	Session           *discordgo.Session
	InteractionCreate *discordgo.InteractionCreate
	RoleCache         InteractionCache[[]*discordgo.Role]
}

func (l LiveInteraction) CommandData() discordgo.ApplicationCommandInteractionData {
	return l.InteractionCreate.ApplicationCommandData()
}

func (l LiveInteraction) Requester() *discordgo.User {
	return l.InteractionCreate.Member.User
}

func (l LiveInteraction) RequesterHasRole(roleName string) (bool, error) {
	role, err := getRoleByName(l, roleName)
	if err != nil {
		return false, err
	}
	for _, roleId := range l.InteractionCreate.Member.Roles {
		if roleId == role.ID {
			return true, nil
		}
	}
	return false, nil
}

func (l LiveInteraction) GetRoles() ([]*discordgo.Role, error) {
	if roles, ok := l.RoleCache.Get(l.InteractionCreate.GuildID); ok {
		return *roles, nil
	}
	return l.Session.GuildRoles(l.InteractionCreate.GuildID)
}

func getRoleByName(l LiveInteraction, name string) (*discordgo.Role, error) {
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

func (l LiveInteraction) AssignRole(userId string, roleName string) error {
	role, err := getRoleByName(l, roleName)
	if err != nil {
		return err
	}
	return l.Session.GuildMemberRoleAdd(l.InteractionCreate.GuildID, userId, role.ID)
}

func (l LiveInteraction) AssignRoleToRequester(roleName string) error {
	return l.AssignRole(l.InteractionCreate.Member.User.ID, roleName)
}

func (l LiveInteraction) RemoveRole(userId string, roleName string) error {
	role, err := getRoleByName(l, roleName)
	if err != nil {
		return err
	}
	return l.Session.GuildMemberRoleRemove(l.InteractionCreate.GuildID, userId, role.ID)
}

func (l LiveInteraction) RemoveRoleFromRequester(roleName string) error {
	return l.RemoveRole(l.InteractionCreate.Member.User.ID, roleName)
}

func (l LiveInteraction) DeferredResponse(msg string, ephemeral bool) error {
	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}
	return l.Session.InteractionRespond(l.InteractionCreate.Interaction, &discordgo.InteractionResponse{
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

func (l LiveInteraction) DeleteChannel(id string) error {
	_, err := l.Session.ChannelDelete(id)
	return err
}

func (l LiveInteraction) EnsureRoleCreated(name string, color int, roles discordgo.Roles) error {
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
		_, err = l.Session.GuildRoleEdit(l.InteractionCreate.GuildID, foundRole.ID, &discordgo.RoleParams{
			Color: &color,
		})
	} else {
		_, err = l.Session.GuildRoleCreate(l.InteractionCreate.GuildID, &discordgo.RoleParams{
			Name:  name,
			Color: &color,
		})
	}
	return err
}
