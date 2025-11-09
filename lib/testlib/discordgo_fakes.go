package testlib

import (
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

func TestDiscordUser() *discordgo.User {
	return &discordgo.User{
		ID:            uuid.NewString(),
		Email:         faker.Email(),
		Username:      faker.Username(),
		Avatar:        "",
		Locale:        "",
		Discriminator: "",
		GlobalName:    faker.Name(),
		Token:         "",
		Verified:      true,
		MFAEnabled:    false,
		Banner:        "",
		AccentColor:   0,
		Bot:           false,
		PublicFlags:   0,
		PremiumType:   0,
		System:        false,
		Flags:         0,
	}
}

func TestDiscordMember(guildId string) *discordgo.Member {
	user := TestDiscordUser()
	daysAgo := rand.Intn(1000) * 24
	return &discordgo.Member{
		GuildID:                    guildId,
		JoinedAt:                   time.Now().UTC().Add(-time.Duration(daysAgo) * time.Hour),
		Nick:                       faker.Name(),
		Deaf:                       false,
		Mute:                       false,
		Avatar:                     "",
		Banner:                     "",
		User:                       user,
		Roles:                      make([]string, 0),
		PremiumSince:               nil,
		Flags:                      0,
		Pending:                    false,
		Permissions:                0,
		CommunicationDisabledUntil: nil,
	}
}
