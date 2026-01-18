package models

import (
	"iter"
	"maps"
	"slices"
)

type Guild struct {
	Id           string `gorm:"primary_key"` // Discord guild ID
	Name         string
	Channels     GuildChannels
	Paused       bool
	GameGoing    bool
	DayNight     bool
	TimeZone     string
	DayTime      *TimeOnly
	NightTime    *TimeOnly
	GameSettings GameSettings
	LastCycleRan string
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
			child := findChannel(appId, slices.Values(*c.Children))
			if child != nil {
				return child
			}
		}
	}
	return nil
}

func (m *Guild) ChannelByAppId(appId string) *GuildChannel {
	return findChannel(appId, maps.Values(m.Channels))
}
