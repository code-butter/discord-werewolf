package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
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
		Global:  true,
		Respond: initServer,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "playing",
			Description: "Signs up for the next round.",
		},
		Global:  true,
		Respond: playing,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "stop_playing",
			Description: "Removes yourself from playing next round.",
		},
		Global:  true,
		Respond: stopPlaying,
	})

	return nil
}

func initServer(i lib.Interaction) error {
	var err error

	if err = i.DeferredResponse(); err != nil {
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
	return nil
}

func playing(i lib.Interaction) error {
	panic("Not yet implemented")
}

func stopPlaying(i lib.Interaction) error {
	panic("Not yet implemented")
}
