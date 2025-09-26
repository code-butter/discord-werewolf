package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/testlib"
	"fmt"
	"iter"
	"maps"
	"slices"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func TestServerInit(t *testing.T) {

	testlib.TestInit()

	guildId := uuid.NewString()
	guildName := "Test Guild"
	i := testlib.NewTestInteraction(guildId, guildName, make([]*discordgo.Channel, 0))
	if err := initServer(i); err != nil {
		t.Fatal(err)
	}
	var guild models.Guild
	if result := lib.GormDB.Where("name = ?", "Test Guild").First(&guild); result.Error != nil {
		t.Fatal(result.Error)
	}
	if guild.Id != guildId {
		t.Errorf("Guild id is %s, expected %s", guildId, guildId)
	}
	if guild.Name != guildName {
		t.Errorf("Guild name is %s, expected %s", guildName, guildName)
	}

	discordChannels, _ := i.Channels()
	dbChannels := maps.Values(guild.Channels)

	for _, ic := range initialChannels {
		var err error
		var parentChannel *models.GuildChannel
		if parentChannel, err = verifyChannels(ic.AppId, dbChannels, discordChannels); err != nil {
			t.Error(err)
			continue
		}
		for _, ch := range *ic.Children {
			if _, err = verifyChannels(ch.AppId, slices.Values(*parentChannel.Children), discordChannels); err != nil {
				t.Error(err)
			}
		}
	}
}

func verifyChannels(appId string, dbChannels iter.Seq[models.GuildChannel], discordChannels []*discordgo.Channel) (*models.GuildChannel, error) {
	var dbChannel *models.GuildChannel
	for ch := range dbChannels {
		if ch.AppId == appId {
			dbChannel = &ch
			break
		}
	}
	if dbChannel == nil {
		return nil, errors.New(fmt.Sprintf("Database channel not found: %s", appId))
	}
	discordChannelFound := false
	for _, dc := range discordChannels {
		if dc.ID == dbChannel.Id {
			discordChannelFound = true
			break
		}
	}
	if !discordChannelFound {
		return nil, errors.New(fmt.Sprintf("Discord channel not found: %s", appId))
	}
	return dbChannel, nil
}
