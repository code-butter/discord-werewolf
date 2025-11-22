package lib

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var timeFormat *regexp.Regexp

func init() {
	timeFormat = regexp.MustCompile(`^[0-9]{2}:[0-9]{2}$`)
}

type GuildSettings struct {
	db    *gorm.DB
	ctx   context.Context
	clock Clock
}

func NewGameSettings(i *do.Injector) (*GuildSettings, error) {
	db, err := do.Invoke[*gorm.DB](i)
	if err != nil {
		return nil, err
	}
	ctx, err := do.Invoke[context.Context](i)
	if err != nil {
		return nil, err
	}
	clock, err := do.Invoke[Clock](i)
	return &GuildSettings{db: db, ctx: ctx, clock: clock}, nil
}

func (gs *GuildSettings) guildRow(guildId string) *gorm.DB {
	// TODO: convert to the generics interface since it supports context cancellation
	return gs.db.Table("guilds").Where("id = ?", guildId)
}

func (gs *GuildSettings) updateJsonValue(guildId string, name string, value string) error {
	row := gs.guildRow(guildId)
	fieldName := fmt.Sprintf("{%s}", name)
	result := row.UpdateColumn("game_settings", datatypes.JSONSet(name).Set(fieldName, value))
	return result.Error
}

func (gs *GuildSettings) StartGame(guildId string) error {
	result := gs.guildRow(guildId).UpdateColumns(map[string]interface{}{
		"game_going":     1,
		"day_night":      0,
		"paused":         0,
		"last_cycle_ran": gs.clock.Now().UTC().Format(time.DateTime),
	})
	// TODO implement next_game_settings and copy settings to game_settings
	return result.Error
}

func (gs *GuildSettings) PauseGame(guildId string) error {
	result := gs.guildRow(guildId).Update("paused", 1)
	return result.Error
}

func (gs *GuildSettings) ResumeGame(guildId string) error {
	result := gs.guildRow(guildId).Update("paused", 0)
	return result.Error
}

func (gs *GuildSettings) EndGame(guildId string) error {
	result := gs.guildRow(guildId).Update("game_going", 0)
	return result.Error
}

func (gs *GuildSettings) SetDayTime(guildId string, time string) error {
	if !timeFormat.MatchString(time) {
		return errors.New("invalid time format (needs HH:MM)")
	}
	result := gs.guildRow(guildId).Update("day_time", time+":00")
	return result.Error
}

func (gs *GuildSettings) SetNightTime(guildId string, time string) error {
	if !timeFormat.MatchString(time) {
		return errors.New("invalid time format (needs HH:MM)")
	}
	result := gs.guildRow(guildId).Update("night_time", time+":00")
	return result.Error
}

func (gs *GuildSettings) SetTimeZone(guildId string, tz string) error {
	if tz != "" {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}
	result := gs.guildRow(guildId).Update("time_zone", tz)
	return result.Error
}

func (gs *GuildSettings) GetTimeZone(guildId string) (*time.Location, error) {
	var tzName string
	result := gs.guildRow(guildId).Pluck("time_zone", &tzName)
	if result.Error != nil {
		return nil, result.Error
	}
	if tzName == "" {
		return SystemTimeZone()
	}
	return time.LoadLocation(tzName)
}

func (gs *GuildSettings) SetDayNight(guildId string, isDay bool) error {
	result := gs.guildRow(guildId).Update("day_night", isDay)
	return result.Error
}
