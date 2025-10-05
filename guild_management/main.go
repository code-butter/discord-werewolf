package guild_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	sets "github.com/hashicorp/go-set/v3"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var initialChannels map[string]models.GuildChannel

func init() {
	initialChannels = map[string]models.GuildChannel{
		"game-instructions": {
			Name:     "Game Instructions",
			AppId:    "game-instructions",
			Children: &[]models.GuildChannel{},
		},
		"the-town": {
			Name:  "The Town",
			AppId: "the-town",
			Children: &[]models.GuildChannel{
				{
					Name:  models.ChannelTownSquare,
					AppId: models.ChannelTownSquare,
				},
				{
					Name:  models.ChannelWerewolves,
					AppId: models.ChannelWerewolves,
				},
				{
					Name:  models.ChannelWitch,
					AppId: models.ChannelWitch,
				},
				{
					Name:  fmt.Sprintf("seer-%s", uuid.New()), // TODO: add more seer channels on demand. Use small, random string instead of UUID
					AppId: "seer-1",
				},
				{
					Name:  fmt.Sprintf("seer-%s", uuid.New()),
					AppId: "seer-2",
				},
				{
					Name:  models.ChannelMasons,
					AppId: models.ChannelMasons,
				},
				{
					Name:  models.ChannelBodyguard,
					AppId: models.ChannelBodyguard,
				},
				{
					Name:  fmt.Sprintf("lovers-%s", uuid.New()), // TODO: add more lovers channels on demand. Use mall, random string instead of UUID
					AppId: "lovers-1",
				},
				{
					Name:  fmt.Sprintf("lovers-%s", uuid.New()),
					AppId: "lovers-2",
				},
				{
					Name:  models.ChannelAfterLife,
					AppId: models.ChannelAfterLife,
				},
			},
		},
		"admin": {
			Name:     "Admin",
			AppId:    "admin",
			Children: &[]models.GuildChannel{},
		},
	}
}

func Setup() error {
	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "init",
			Description: "Initializes the server. Wipes out any data previously stored.",
		},
		Global:      true,
		Respond:     initServer,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Pings the server. Responds with 'pong'.",
		},
		Global:  true,
		Respond: ping,
	})

	tzs := lib.AllTimeZoneNames()
	locationMatcher := regexp.MustCompile(`^([^/])+`)
	locationSet := sets.New[string](0)
	for _, tz := range tzs {
		area := locationMatcher.FindString(tz)
		if area != "" {
			locationSet.Insert(area)
		}
	}
	var locationChoices []*discordgo.ApplicationCommandOptionChoice
	for tzLocation := range locationSet.Items() {
		if tzLocation == "Etc" {
			continue
		}
		locationChoices = append(locationChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  tzLocation,
			Value: tzLocation,
		})
	}

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "get_timezones",
			Description: "Get timezones for the server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "area",
					Description: "Get timezones in this general area.",
					Required:    true,
					Choices:     locationChoices,
				},
			},
		},
		Global:      true,
		Respond:     getTimeZones,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "set_timezone",
			Description: "Sets the timezone for the server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timezone",
					Description: "Sets the timezone for the server.",
					Required:    true,
				},
			},
		},
		Global:      true,
		Respond:     setTimeZone,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	return nil
}

func getTimeZones(ia *lib.InteractionArgs) error {
	var err error
	if err = ia.Interaction.DeferredResponse("Loading timezones...", true); err != nil {
		return err
	}
	areaName := ia.Interaction.CommandData().GetOption("area").Value.(string)
	matcher := regexp.MustCompile(fmt.Sprintf("^%s/", areaName))
	tzs := lib.AllTimeZoneNames()
	builder := strings.Builder{}
	for _, tz := range tzs {
		if matcher.MatchString(tz) {
			builder.WriteString(fmt.Sprintln(tz))
			if builder.Len() > 1900 { // discord max message length is 2000 characters
				if err = ia.Interaction.FollowupMessage(builder.String(), true); err != nil {
					return err
				}
				builder.Reset()
			}
		}
	}
	return ia.Interaction.FollowupMessage(builder.String(), true)
}

func setTimeZone(ia *lib.InteractionArgs) error {
	data := ia.Interaction.CommandData()
	tzName := data.GetOption("timezone").Value.(string)
	_, err := time.LoadLocation(tzName)
	if err != nil {
		_ = ia.Interaction.Respond("Unable to set timezone", true)
		return err
	}
	_, err = ia.DB.ExecContext(ia.Ctx, "UPDATE guilds SET time_zone = ? WHERE id = ?", tzName, ia.Interaction.GuildId())
	if err != nil {
		_ = ia.Interaction.Respond("Unable to set timezone", true)
		return err
	}
	_ = ia.Interaction.Respond(fmt.Sprintf("Set timezone to %s", tzName), true)
	return nil
}

func ping(ia *lib.InteractionArgs) error {
	return ia.Interaction.Respond("Pong!", false)
}

func initServer(ia *lib.InteractionArgs) error {
	var err error

	if err = ia.Interaction.DeferredResponse("Initializing server...", true); err != nil {
		return errors.Wrap(err, "Could not send deferred response to Discord")
	}

	defer func() {
		if err != nil {
			_ = ia.Interaction.FollowupMessage("There was an error setting up the server.", true)
		} else {
			_ = ia.Interaction.FollowupMessage("Server initialized!", true)
		}
	}()

	guild, err := ia.Session.Guild()
	if err != nil {
		return errors.Wrap(err, "Could not get current guild")
	}

	var guildRecord *models.Guild
	if result := ia.GormDB.Where("id = ?", guild.ID).First(&guildRecord); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			guildRecord = nil
		} else {
			return errors.Wrap(result.Error, "Could not get guild record")
		}
	}
	if guildRecord == nil {
		guildRecord = &models.Guild{
			Name: guild.Name,
			Id:   guild.ID,
		}
	}

	////////////////////////////////////////
	// TODO: REMOVE/MODIFY THIS BEFORE PROD DEPLOY
	discordChannels, err := ia.Session.Channels()

	for _, channel := range discordChannels {
		if channel.Name == "general" {
			continue
		}
		if err = ia.Session.DeleteChannel(channel.ID); err != nil {
			return errors.Wrap(err, "Could not delete channel")
		}
	}
	////////////////////////////////////////

	roles, err := ia.Session.GetRoles()
	if err != nil {
		return errors.Wrap(err, "Could not get guild roles")
	}
	roleMap := make(map[string]*discordgo.Role)
	for _, role := range roles {
		roleMap[role.Name] = role
	}

	saveChannels := models.GuildChannels{}
	for _, initChannel := range initialChannels {
		discordCat, err := ia.Session.CreateCategoryChannel(initChannel.Name)
		if err != nil {
			return errors.Wrap(err, "Could not create category channel")
		}
		if initChannel.AppId == "admin" {
			if err = ia.Session.RoleChannelPermissions(discordCat.ID, roleMap["Admin"].ID, discordgo.PermissionViewChannel, 0); err != nil {
				return errors.Wrap(err, "Could not set channel permissions for admin role")
			}
			if err = ia.Session.RoleChannelPermissions(discordCat.ID, guild.ID, 0, discordgo.PermissionViewChannel); err != nil {
				return errors.Wrap(err, "Could not set channel permissions for admin role")
			}
		}
		cat := models.GuildChannel{
			Name:     initChannel.Name,
			Id:       discordCat.ID,
			AppId:    initChannel.AppId,
			Children: &[]models.GuildChannel{},
		}
		var catChildren []models.GuildChannel
		for _, child := range *initChannel.Children {
			discordChannel, err := ia.Session.CreateTextChannel(child.Name, cat.Id)
			if err != nil {
				msg := fmt.Sprintf("Could not create text channel for child %s in guild %s", child.Name, guild.Name)
				return errors.Wrap(err, msg)
			}
			if err = ia.Session.RoleChannelPermissions(discordChannel.ID, guild.ID, 0, discordgo.PermissionSendMessages); err != nil {
				return errors.Wrap(err, "Could not set universal channel permissions for channel "+child.Name)
			}
			if child.AppId == models.ChannelTownSquare {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Alive"].ID, discordgo.PermissionSendMessages|discordgo.PermissionViewChannel, 0); err != nil {
					return errors.Wrap(err, "Could not set Alive channel permissions for channel "+child.Name)
				}
			} else {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Alive"].ID, 0, discordgo.PermissionViewChannel); err != nil {
					return errors.Wrap(err, "Could not set Alive channel permissions for channel "+child.Name)
				}
			}
			if child.AppId == models.ChannelAfterLife {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Dead"].ID, discordgo.PermissionSendMessages|discordgo.PermissionViewChannel, 0); err != nil {
					return errors.Wrap(err, "Could not set Dead channel permissions for channel "+child.Name)
				}
			}
			catChildren = append(catChildren, models.GuildChannel{
				Name:  child.Name,
				AppId: child.AppId,
				Id:    discordChannel.ID,
			})
		}
		cat.Children = &catChildren
		saveChannels[cat.AppId] = cat
	}
	guildRecord.Channels = saveChannels

	result := ia.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "channels"}),
	}).Create(&guildRecord)
	if result.Error != nil {
		err = result.Error
		return errors.Wrap(result.Error, "Could not update guild record")
	}

	guildRoles, err := ia.Session.GetRoles()
	if err != nil {
		return errors.Wrap(err, "Could not get guild roles")
	}
	for _, role := range lib.Roles {
		if err = ia.Session.EnsureRoleCreated(role.Name, role.Color, guildRoles); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not create role %s", role.Name))
		}
	}
	return nil
}
