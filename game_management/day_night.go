package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/listeners"
	"discord-werewolf/lib/models"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func StartDay(guildId string, s lib.SessionArgs) error {
	var err error
	var result *gorm.DB
	var guild models.Guild
	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	if result = gormDB.Where("id = ?", guildId).First(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Guild not found")
	}
	err = listeners.DayStartListeners.Trigger(&s, listeners.DayStartData{
		Guild: guild,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to start day from triggers")
	}
	guild.DayNight = true
	if result = gormDB.Save(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Could not save guild")
	}

	return nil
}

func StartNight(guildId string, s lib.SessionArgs) error {
	var err error
	var result *gorm.DB
	var guild models.Guild

	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	if result = gormDB.Where("id = ?", guildId).First(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Guild not found")
	}
	err = listeners.NightStartListeners.Trigger(&s, listeners.NightStartData{
		Guild: guild,
	})
	if err != nil {
		return errors.Wrap(err, "Failed to start night from triggers")
	}
	guild.DayNight = false
	if result = gormDB.Save(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Could not save guild")
	}
	return nil
}

func TimedDayNight(i *do.Injector) {
	var err error

	clock := do.MustInvoke[lib.Clock](i)
	ctx := do.MustInvoke[context.Context](i)
	gormDB := do.MustInvoke[*gorm.DB](i)

	systemTz, err := lib.SystemTimeZone()

	if err != nil {
		log.Println(err)
		systemTz = time.UTC
	}

	for ctx.Err() == nil {
		cutoff := clock.Now().UTC().Add(-23*time.Hour + 59*time.Minute)
		now := clock.Now().UTC()
		var guilds []models.Guild
		if result := gormDB.Where("game_going = 1 AND paused = 0").Find(&guilds); result.Error != nil {
			panic(result.Error)
		}
		var finishedGuildIds []string
		for _, guild := range guilds {
			lastCycleRan := guild.LastCycleRan
			if lastCycleRan == nil {
				lastCycleRan = &now
			}
			sa := lib.SessionArgs{
				Session:  lib.GetGuildDiscordSession(guild.Id),
				Injector: i,
			}
			if lastCycleRan.Before(cutoff) {
				if guild.DayNight {
					err = StartNight(guild.Id, sa)
				} else {
					err = StartDay(guild.Id, sa)
				}
				if err != nil {
					log.Println(err)
					continue
				}
				finishedGuildIds = append(finishedGuildIds, guild.Id)
				continue
			}
			var guildTz *time.Location
			if guild.TimeZone == "" {
				guildTz = systemTz
			} else {
				guildTz, err = time.LoadLocation(guild.TimeZone)
				if err != nil {
					log.Println(err)
					guildTz = systemTz
				}
			}
			if guild.DayNight {
				nightTime := guild.NightTime
				if nightTime == nil {
					newNightTime := time.Date(0, 0, 0, 18, 0, 0, 0, guildTz).UTC()
					nightTime = &models.TimeOnly{Time: &newNightTime}
				}
				if nightTime.BeforeOrOn(now) && nightTime.AfterOrOn(*lastCycleRan) {
					if err = StartNight(guild.Id, sa); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			} else {
				dayTime := guild.DayTime
				if dayTime == nil {
					newDayTime := time.Date(0, 0, 0, 6, 0, 0, 0, guildTz).UTC()
					dayTime = &models.TimeOnly{Time: &newDayTime}
				}
				if dayTime.BeforeOrOn(now) && dayTime.AfterOrOn(*lastCycleRan) {
					if err = StartDay(guild.Id, sa); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			}
		}
		_, err = gorm.G[models.Guild](gormDB).
			Where("id in ?", finishedGuildIds).
			Update(ctx, "last_cycle_ran", now)
		if err != nil {
			lib.Fatal(err)
		}
		time.Sleep(time.Second * 15)
	}
}

func triggerNight(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering night...", true)
	if err != nil {
		return err
	}
	if err = StartNight(args.Interaction.GuildId(), args.SessionArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Night triggered!", true)
}

func triggerDay(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering day...", true)
	if err != nil {
		return err
	}
	if err = StartDay(args.Interaction.GuildId(), args.SessionArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Day triggered!", true)
}

func nightListener(s *lib.SessionArgs, data listeners.NightStartData) error {
	var err error
	var result *gorm.DB
	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	ctx := do.MustInvoke[context.Context](s.Injector)
	townChannel := data.Guild.ChannelByAppId(models.ChannelTownSquare)
	if townChannel == nil {
		return errors.New("No town channel found for guild " + data.Guild.Id)
	}
	var voted string
	result = gormDB.
		Model(&models.GuildVote{}).
		Select("voting_for_id").
		Group("voting_for_id").
		Order("COUNT(*) DESC").
		Where("guild_id = ?", data.Guild.Id).
		Limit(1).
		Pluck("voting_for_id", &voted)
	if result.Error != nil {
		msg := fmt.Sprintf("Could not get votes for guild %s with ID %s", data.Guild.Name, data.Guild.Id)
		return errors.Wrap(result.Error, msg)
	}

	if voted == "" {
		msg := &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "No one voted.",
			Description: "No one was hanged.",
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
	} else {
		var character models.GuildCharacter
		result = gormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).First(&character)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Could not find character with ID "+voted)
		}
		if err = s.Session.RemoveRole(voted, "Alive"); err != nil {
			return errors.Wrap(err, "Could not remove Alive role")
		}
		if err = s.Session.AssignRole(voted, "Dead"); err != nil {
			return errors.Wrap(err, "Could not assign Dead role")
		}
		msg := &discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeRich,
			Title: fmt.Sprintf("The town has hanged <@%s>.", voted),
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
		character.ExtraData["death_cause"] = "hanged"
		if result = gormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).Save(&character); result.Error != nil {
			return errors.Wrap(result.Error, "Could not update character with ID "+voted)
		}
	}
	_, err = gorm.G[models.GuildVote](gormDB).Where("guild_id = ?", data.Guild.Id).Delete(ctx)
	return err
}
