package models

import (
	"iter"
	"maps"
	"slices"
	"time"
)

type Guild struct {
	Id           string `gorm:"primary_key"`
	Name         string
	Channels     GuildChannels
	Paused       bool
	GameGoing    bool
	DayNight     bool
	TimeZone     string
	DayTime      *TimeOnly
	NightTime    *TimeOnly
	GameSettings JsonMap
	LastCycleRan *time.Time `gorm:"type:datetime"`
}

func findChannel(appId string, channels iter.Seq[GuildChannel]) *GuildChannel {
	if channels == nil {
		return nil
	}
	for c := range channels {
		if c.AppId == appId {
			return &c
		}
		if c.Children != nil {
			return findChannel(appId, slices.Values(*c.Children))
		}
	}
	return nil
}

func (m *Guild) ChannelByAppId(appId string) *GuildChannel {
	return findChannel(appId, maps.Values(m.Channels))
}

func (m *Guild) Set(settingName string, value interface{}) {
	if m.GameSettings == nil {
		m.GameSettings = JsonMap{}
	}
	m.GameSettings[settingName] = value
}

func (m *Guild) Get(settingName string) interface{} {
	if m.GameSettings == nil {
		m.GameSettings = JsonMap{}
	}
	return m.GameSettings[settingName]
}
