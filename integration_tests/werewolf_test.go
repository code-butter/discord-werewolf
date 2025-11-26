package integration_tests

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/shared"
	"discord-werewolf/lib/testlib"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func TestWerewolfKill(t *testing.T) {
	s := StartDefaultIntegratedTestGame(15, 15)

	characters, err := s.GuildCharacters()
	if err != nil {
		t.Fatalf("Error getting guild's characters: %v", err)
	}
	wolves, villagers := getWolvesVillagers(characters)

	if len(wolves) != 3 {
		t.Fatalf("Werewolves do not equal required amount")
	}

	toDie := villagers[0]
	toNotDie := villagers[1]
	for i, wolf := range wolves {
		if i%3 == 0 {
			err = testVoteToKill(s, wolf.Id, toNotDie.Id)
		} else {
			err = testVoteToKill(s, wolf.Id, toDie.Id)
		}
		if err != nil {
			t.Fatal(err)
		}
	}

	if err = shared.StartDay(s); err != nil {
		t.Fatal(err)
	}

	aliveMembers, err := s.Session.GuildMembersWithRole(lib.RoleAlive)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, member := range aliveMembers {
		if member.User.ID == toNotDie.Id {
			found = true
			break
		}
	}
	if !found {
		t.Error("Least voted player is not alive")
	}
	deadMembers, err := s.Session.GuildMembersWithRole(lib.RoleDead)
	if err != nil {
		t.Fatal(err)
	}
	found = false
	for _, member := range deadMembers {
		if member.User.ID == toDie.Id {
			found = true
			break
		}
	}
	if !found {
		t.Error("Most voted player is not dead")
	}
}

func TestWerewolfGameEnd(t *testing.T) {
	s := StartDefaultIntegratedTestGame(5, 5)
	characters, err := s.GuildCharacters()
	if err != nil {
		t.Fatalf("Error getting guild's characters: %v", err)
	}
	wolves, villagers := getWolvesVillagers(characters)

	if len(wolves) != 1 {
		t.Fatalf("Werewolves do not equal required amount")
	}

	for i := 0; i < len(villagers)-1; i++ {
		aliveMembers, err := s.Session.GuildMembersWithRole(lib.RoleAlive)
		if err != nil {
			t.Fatal(err)
		}
		var toKill string
		for _, member := range aliveMembers {
			if member.User.ID != wolves[0].Id {
				toKill = member.User.ID
				break
			}
		}
		if err = testVoteToKill(s, wolves[0].Id, toKill); err != nil {
			t.Fatal(err)
		}
		if err = shared.StartDay(s); err != nil {
			t.Fatal(err)
		}
		if err = shared.StartNight(s); err != nil {
			t.Fatal(err)
		}
	}

	guild, err := s.AppGuild()
	if err != nil {
		t.Fatal(err)
	}
	if guild.GameGoing {
		t.Error("Game should be ended")
	}

}

func getWolvesVillagers(characters []*models.GuildCharacter) (wolves, villagers []*models.GuildCharacter) {
	for _, character := range characters {
		if character.CharacterId == models.CharacterWolf || character.CharacterId == models.CharacterWolfCub {
			wolves = append(wolves, character)
		} else {
			villagers = append(villagers, character)
		}
	}
	return
}

func testVoteToKill(s *lib.SessionArgs, wolfId, characterId string) error {
	guild, err := s.AppGuild()
	if err != nil {
		return err
	}
	member, err := s.Session.GuildMember(wolfId)
	if err != nil {
		return err
	}
	channel := guild.ChannelByAppId(models.ChannelWerewolves)
	if channel == nil {
		return errors.New("wolf channel not found")
	}
	CallInteraction(s, testlib.TestInteractionOptions{
		Requester: member.User,
		CommandData: discordgo.ApplicationCommandInteractionData{
			Name: "kill",
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  "user",
					Value: characterId,
				},
			},
		},
		ChannelId: channel.Id,
	})
	return nil
}
