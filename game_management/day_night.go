package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/shared"
	"fmt"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func TimedDayNight(i *do.Injector, loopSleep time.Duration) {
	var err error

	clock := do.MustInvoke[lib.Clock](i)
	ctx := do.MustInvoke[context.Context](i)
	gormDB := do.MustInvoke[*gorm.DB](i)
	sessionProvider := do.MustInvoke[lib.DiscordSessionProvider](i)

	systemTz, err := lib.SystemTimeZone()

	if err != nil {
		log.Println(err)
		systemTz = time.UTC
	}

	for ctx.Err() == nil {
		now := clock.Now().UTC()
		cutoff := now.Add(-24 * time.Hour)

		var guilds []models.Guild
		if result := gormDB.Where("game_going = 1 AND paused = 0").Find(&guilds); result.Error != nil {
			log.Printf("Could not get guilds: %v", result.Error)
		}
		var finishedGuildIds []string
		for _, guild := range guilds {
			sa := &lib.SessionArgs{
				Session:  sessionProvider.GetSession(guild.Id),
				Injector: i,
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
			guildNow := now.In(guildTz)

			var lastCycleRan time.Time
			lastCycleRan, err = time.ParseInLocation(time.DateTime, guild.LastCycleRan, time.UTC)
			if err != nil {
				log.Printf("Error parsing last cycle time: %s\n", err)
				continue
			}

			lastCycleRan = lastCycleRan.In(guildTz)

			// If it's been longer than a day
			if lastCycleRan.Before(cutoff) {
				if guild.DayNight {
					if err = shared.StartNight(sa); err != nil {
						log.Println(err)
						continue
					}
				} else {
					if err = shared.StartDay(sa); err != nil {
						log.Println(err)
						continue
					}
				}
				finishedGuildIds = append(finishedGuildIds, guild.Id)
				continue
			}

			guildNightTime := guild.NightTime
			if guildNightTime == nil {
				guildNightTime = models.NewTimeOnly(18, 0, 0)
			}
			guildDayTime := guild.DayTime
			if guildDayTime == nil {
				guildDayTime = models.NewTimeOnly(6, 0, 0)
			}
			nightTime := guildNightTime.TimeOnDate(guildNow)
			dayTime := guildDayTime.TimeOnDate(guildNow)

			if guild.DayNight {
				if lastCycleRan.Before(nightTime) && guildNow.After(nightTime) {
					if err = shared.StartNight(sa); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			} else {
				if lastCycleRan.Before(dayTime) && guildNow.After(dayTime) {
					if err = shared.StartDay(sa); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			}
		}
		_, err = gorm.G[models.Guild](gormDB).
			Where("id in ?", finishedGuildIds).
			Update(ctx, "last_cycle_ran", now.Format(time.DateTime))
		if err != nil {
			lib.Fatal(err)
		}
		time.Sleep(loopSleep)
	}
}

func triggerNight(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering night...", true)
	if err != nil {
		return err
	}
	if err = shared.StartNight(args.SessionArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Night triggered!", true)
}

func triggerDay(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering day...", true)
	if err != nil {
		return err
	}
	if err = shared.StartDay(args.SessionArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Day triggered!", true)
}

func nightListener(s *lib.SessionArgs, data lib.NightStartData) error {
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
		character, err := s.GuildCharacter(voted)
		if err != nil {
			return errors.Wrap(err, "Could not get character "+voted)
		}
		msg := &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "The angry mob demands justice",
			Description: fmt.Sprintf("<@%s> has been hanged", voted),
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
		return shared.KillCharacter(s, character, "hanged")
	}
	_, err = gorm.G[models.GuildVote](gormDB).Where("guild_id = ?", data.Guild.Id).Delete(ctx)
	return err
}
