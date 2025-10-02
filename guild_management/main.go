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
					Name:  "town-square",
					AppId: "town-square",
				},
				{
					Name:  "werewolves",
					AppId: "werewolves",
				},
				{
					Name:  "witch",
					AppId: "witch",
				},
				{
					Name:  fmt.Sprintf("seer-%s", uuid.New()), // TODO: add more seer channels on demand
					AppId: "seer-1",
				},
				{
					Name:  fmt.Sprintf("seer-%s", uuid.New()),
					AppId: "seer-2",
				},
				{
					Name:  "masons",
					AppId: "masons",
				},
				{
					Name:  "bodyguard",
					AppId: "bodyguard",
				},
				{
					Name:  fmt.Sprintf("lovers-%s", uuid.New()), // TODO: add more lovers channels on demand
					AppId: "lovers-1",
				},
				{
					Name:  fmt.Sprintf("lovers-%s", uuid.New()),
					AppId: "lovers-2",
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

func getTimeZones(i lib.Interaction) error {
	var err error
	if err = i.DeferredResponse("Loading timezones...", true); err != nil {
		return err
	}
	areaName := i.CommandData().GetOption("area").Value.(string)
	matcher := regexp.MustCompile(fmt.Sprintf("^%s/", areaName))
	tzs := lib.AllTimeZoneNames()
	builder := strings.Builder{}
	for _, tz := range tzs {
		if matcher.MatchString(tz) {
			builder.WriteString(fmt.Sprintln(tz))
			if builder.Len() > 1900 { // discord max message length is 2000 characters
				if err = i.FollowupMessage(builder.String(), true); err != nil {
					return err
				}
				builder.Reset()
			}
		}
	}
	return i.FollowupMessage(builder.String(), true)
}

func setTimeZone(i lib.Interaction) error {
	data := i.CommandData()
	tzName := data.GetOption("timezone").Value.(string)
	_, err := time.LoadLocation(tzName)
	if err != nil {
		_ = i.Respond("Unable to set timezone", true)
		return err
	}
	_, err = lib.DB.ExecContext(lib.Ctx, "UPDATE guilds SET time_zone = ? WHERE id = ?", tzName, i.GuildId())
	if err != nil {
		_ = i.Respond("Unable to set timezone", true)
		return err
	}
	_ = i.Respond(fmt.Sprintf("Set timezone to %s", tzName), true)
	return nil
}

func ping(i lib.Interaction) error {
	return i.Respond("Pong!", false)
}

func initServer(i lib.Interaction) error {
	var err error

	if err = i.DeferredResponse("Initializing server...", true); err != nil {
		return errors.Wrap(err, "Could not send deferred response to Discord")
	}

	defer func() {
		if err != nil {
			_ = i.FollowupMessage("There was an error setting up the server.", true)
		} else {
			_ = i.FollowupMessage("Server initialized!", true)
		}
	}()

	guild, err := i.Guild()
	if err != nil {
		return errors.Wrap(err, "Could not get current guild")
	}

	var guildRecord *models.Guild
	if result := lib.GormDB.Where("id = ?", guild.ID).First(&guildRecord); result.Error != nil {
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
	discordChannels, err := i.Channels()

	for _, channel := range discordChannels {
		if err = i.DeleteChannel(channel.ID); err != nil {
			return errors.Wrap(err, "Could not delete channel")
		}
	}
	////////////////////////////////////////

	saveChannels := models.GuildChannels{}
	for _, initChannel := range initialChannels {
		discordCat, err := i.CreateCategoryChannel(initChannel.Name)
		if err != nil {
			return errors.Wrap(err, "Could not create category channel")
		}
		cat := models.GuildChannel{
			Name:     initChannel.Name,
			Id:       discordCat.ID,
			AppId:    initChannel.AppId,
			Children: &[]models.GuildChannel{},
		}
		var catChildren []models.GuildChannel
		for _, child := range *initChannel.Children {
			discordChannel, err := i.CreateTextChannel(child.Name, cat.Id)
			if err != nil {
				msg := fmt.Sprintf("Could not create text channel for child %s in guild %s", child.Name, guild.Name)
				return errors.Wrap(err, msg)
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

	result := lib.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "channels"}),
	}).Create(&guildRecord)
	if result.Error != nil {
		err = result.Error
		return errors.Wrap(result.Error, "Could not update guild record")
	}

	guildRoles, err := i.GetRoles()
	if err != nil {
		return errors.Wrap(err, "Could not get guild roles")
	}
	for _, role := range lib.Roles {
		if err = i.EnsureRoleCreated(role.Name, role.Color, guildRoles); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not create role %s", role.Name))
		}
	}
	return nil
}
