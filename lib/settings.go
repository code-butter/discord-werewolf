package lib

import (
	"context"
	"discord-werewolf/lib/models"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/do"
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

func GameSettingsProvider(i *do.Injector) (*GuildSettings, error) {
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

func (gs *GuildSettings) guildRow(ctx context.Context, guildId string) *gorm.DB {
	return gs.db.WithContext(ctx).Table("guilds").Where("id = ?", guildId)
}

func (gs *GuildSettings) StartGame(ctx context.Context, guildId string) error {
	result := gs.guildRow(ctx, guildId).UpdateColumns(map[string]interface{}{
		"game_going":     1,
		"day_night":      0,
		"paused":         0,
		"last_cycle_ran": gs.clock.Now().UTC().Format(time.DateTime),
	})
	return result.Error
}

func (gs *GuildSettings) PauseGame(ctx context.Context, guildId string) error {
	result := gs.guildRow(ctx, guildId).Update("paused", 1)
	return result.Error
}

func (gs *GuildSettings) ResumeGame(ctx context.Context, guildId string) error {
	result := gs.guildRow(ctx, guildId).Update("paused", 0)
	return result.Error
}

func (gs *GuildSettings) EndGame(ctx context.Context, guildId string) error {
	result := gs.guildRow(ctx, guildId).Update("game_going", 0)
	return result.Error
}

func (gs *GuildSettings) SetDayTime(ctx context.Context, guildId string, time string) error {
	if !timeFormat.MatchString(time) {
		return errors.New("invalid time format (needs HH:MM)")
	}
	result := gs.guildRow(ctx, guildId).Update("day_time", time+":00")
	return result.Error
}

func (gs *GuildSettings) SetNightTime(ctx context.Context, guildId string, time string) error {
	if !timeFormat.MatchString(time) {
		return errors.New("invalid time format (needs HH:MM)")
	}
	result := gs.guildRow(ctx, guildId).Update("night_time", time+":00")
	return result.Error
}

func (gs *GuildSettings) SetTimeZone(ctx context.Context, guildId string, tz string) error {
	if tz != "" {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}
	result := gs.guildRow(ctx, guildId).Update("time_zone", tz)
	return result.Error
}

func (gs *GuildSettings) GetTimeZone(ctx context.Context, guildId string) (*time.Location, error) {
	var tzName string
	result := gs.guildRow(ctx, guildId).Pluck("time_zone", &tzName)
	if result.Error != nil {
		return nil, result.Error
	}
	if tzName == "" {
		return SystemTimeZone()
	}
	return time.LoadLocation(tzName)
}

func (gs *GuildSettings) SetDayNight(ctx context.Context, guildId string, isDay bool) error {
	result := gs.guildRow(ctx, guildId).Update("day_night", isDay)
	return result.Error
}

func (gs *GuildSettings) UpdateSettings(ctx context.Context, guildId string, settings models.GameSettings) error {
	result := gs.guildRow(ctx, guildId).Update("game_settings", settings)
	return result.Error
}
