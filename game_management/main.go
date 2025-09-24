package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"gorm.io/gorm/clause"
)

func Init() error {

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

	var guildRecord models.Guild
	if result := lib.GormDB.Select(&guildRecord, "id = ?", guild.ID); result.Error != nil {
		return errors.Wrap(result.Error, "Could not get guild record")
	}

	discordChannels, err := i.Channels()
	if err != nil {
		return errors.Wrap(err, "Could not get discord channels")
	}

	discordChannelNames := map[string]*discordgo.Channel{}
	discordChannelIds := map[string]*discordgo.Channel{}
	recordChannelNames := map[string]*models.GuildChannel{}
	recordChannelIds := map[string]*models.GuildChannel{}
	recordChannelAppIds := map[string]*models.GuildChannel{}

	initialChannels := []models.GuildChannel{
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
	}

	for _, channel := range discordChannels {
		discordChannelNames[channel.Name] = channel
		discordChannelIds[channel.ID] = channel
	}

	if guildRecord.Channels == nil {
		guildRecord.Channels = make([]models.GuildChannel, 0)
	}

	for _, channel := range guildRecord.Channels {
		recordChannelNames[channel.Name] = &channel
		recordChannelIds[channel.Id] = &channel
		recordChannelAppIds[channel.Id] = &channel
	}

	for _, ic := range initialChannels {
		record, recordFound := recordChannelAppIds[ic.AppId]
		if recordFound && record.Id != "" {
			if _, ok := discordChannelIds[record.Id]; !ok {
				if discordChannel, ok2 := discordChannelNames[record.Name]; ok2 {
					record.Id = discordChannel.ID
				} else {
					newChannel, err := i.CreateChannel(ic.Name)
					if err != nil {
						return errors.Wrap(err, "Could not create channel")
					}
					record.Id = newChannel.ID
					record.Name = newChannel.Name
				}
			}
		} else {
			newChannel, err := i.CreateChannel(ic.Name)
			if err != nil {
				return errors.Wrap(err, "Could not create channel")
			}
			guildRecord.Channels = append(guildRecord.Channels, models.GuildChannel{
				Name:  newChannel.Name,
				Id:    newChannel.ID,
				AppId: ic.AppId,
			})
		}
	}

	result := lib.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "channels"}),
	}).Create(&guildRecord)
	if result.Error != nil {
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
