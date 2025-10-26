package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/listeners"
	"discord-werewolf/lib/models"
	"log"
	"time"

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
