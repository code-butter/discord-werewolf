package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func checkWinConditions(s *lib.SessionArgs, data lib.CharacterDeathData) error {
	alive, err := s.Session.GuildMembersWithRole(lib.RoleAlive)
	if err != nil {
		return err
	}

	townChannel := data.Guild.ChannelByAppId(models.ChannelTownSquare)
	if townChannel == nil {
		return errors.New("No town channel found for guild " + data.Guild.Id)
	}

	aliveIds := make([]string, 0, len(alive))
	for _, m := range alive {
		aliveIds = append(aliveIds, m.User.ID)
	}
	characters, err := s.GuildCharacters()
	if err != nil {
		return err
	}
	aliveVillagers := make([]*models.GuildCharacter, 0)
	deadVillagers := make([]*models.GuildCharacter, 0)
	aliveWolves := make([]*models.GuildCharacter, 0)
	deadWolves := make([]*models.GuildCharacter, 0)

	for _, c := range characters {
		if c.CharacterId == models.CharacterWolf || c.CharacterId == models.CharacterWolfCub {
			if slices.Contains(aliveIds, c.Id) {
				aliveWolves = append(aliveWolves, c)
			} else {
				deadWolves = append(deadWolves, c)
			}
		} else {
			if slices.Contains(aliveIds, c.Id) {
				aliveVillagers = append(aliveVillagers, c)
			} else {
				deadVillagers = append(deadVillagers, c)
			}
		}
	}

	var msg *discordgo.MessageEmbed
	aliveWolfCount := len(aliveWolves)
	aliveVillagerCount := len(aliveVillagers)
	if aliveWolfCount == 0 {
		memberList := aliveAndDeadList(aliveVillagers, deadVillagers)
		msg = &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Villagers win!",
			Description: "There is no more evil in the town.\n\n" + memberList,
		}
	} else if aliveWolfCount >= aliveVillagerCount {
		memberList := aliveAndDeadList(aliveWolves, deadWolves)
		msg = &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Werewolves win!",
			Description: "They now outnumber the villagers.\n\n" + memberList,
		}
	}
	if msg != nil {
		return lib.GameOver{MessageEmbed: *msg}
	}
	return nil
}

func aliveAndDeadList(alive, dead []*models.GuildCharacter) (memberList string) {
	memberList = "Alive:"
	for _, v := range alive {
		memberList += "\n  " + "<@" + v.Id + "> - " + v.CharacterDescription()
	}
	memberList += "\n\nDead:"
	for _, v := range dead {
		memberList += "\n  <@" + v.Id + "> - " + v.CharacterDescription()
	}
	if len(dead) == 0 {
		memberList += "\n  (none)"
	}
	return
}
