package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

func StartDay(guildId string, session *discordgo.Session) error {

	return nil
}

func StartNight(guildId string, session *discordgo.Session) error {

	return nil
}

func TimedDayNight(ctx context.Context, session *discordgo.Session, gormDB *gorm.DB) {
	var err error

	systemTz, err := lib.SystemTimeZone()

	if err != nil {
		log.Println(err)
		systemTz = time.UTC
	}

	for ctx.Err() == nil {
		cutoff := time.Now().UTC().Add(-23*time.Hour + 59*time.Minute)
		now := time.Now().UTC()
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
			if lastCycleRan.Before(cutoff) {
				if guild.DayNight {
					err = StartNight(guild.Id, session)
				} else {
					err = StartDay(guild.Id, session)
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
					if err = StartNight(guild.Id, session); err != nil {
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
					if err = StartDay(guild.Id, session); err != nil {
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
		time.Sleep(time.Minute * 5)
	}
}
