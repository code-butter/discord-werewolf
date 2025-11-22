package integration_tests

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/shared"
	"discord-werewolf/lib/testlib"
	"math/rand"
	"slices"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func TestTownHanging(t *testing.T) {
	s := StartDefaultIntegratedTestGame(10, 8)
	if err := shared.StartDay(s); err != nil {
		t.Fatal(err)
	}
	characters, err := s.GuildCharacters()
	if err != nil {
		t.Fatal(err)
	}
	var selectedToDie []string
	toDie := randomCharacter(&selectedToDie, characters)
	toNotDie1 := randomCharacter(&selectedToDie, characters)
	toNotDie2 := randomCharacter(&selectedToDie, characters)
	majority := characters[0:3]
	minority1 := characters[4:6]
	minority2 := characters[7:7]
	for _, character := range majority {
		if err := testVoteToHang(s, character.Id, toDie.Id); err != nil {
			t.Fatal(err)
		}
	}
	for _, character := range minority1 {
		if err := testVoteToHang(s, character.Id, toNotDie1.Id); err != nil {
			t.Fatal(err)
		}
	}
	for _, character := range minority2 {
		if err := testVoteToHang(s, character.Id, toNotDie2.Id); err != nil {
			t.Fatal(err)
		}
	}
	if err := shared.StartNight(s); err != nil {
		t.Fatal(err)
	}

	aliveMembers, err := s.Session.GuildMembersWithRole(lib.RoleAlive)
	if err != nil {
		t.Fatal(err)
	}
	deadMembers, err := s.Session.GuildMembersWithRole(lib.RoleDead)
	if err != nil {
		t.Fatal(err)
	}
	if len(deadMembers) != 1 || deadMembers[0].User.ID != toDie.Id {
		t.Error("error with lynching villager")
	}
	if len(aliveMembers) != len(characters)-1 {
		t.Error("incorrect count of alive players")
	}
}

func randomCharacter(selectedIds *[]string, characters []*models.GuildCharacter) *models.GuildCharacter {
	if len(*selectedIds) >= len(characters) {
		panic("Cannot get random character. Length of selected is bigger than available characters")
	}
	var character *models.GuildCharacter
	for character == nil || slices.Contains(*selectedIds, character.Id) {
		character = characters[rand.Intn(len(characters))]
	}
	newSlice := append(*selectedIds, character.Id)
	selectedIds = &newSlice
	return character
}

func testVoteToHang(s lib.SessionArgs, voterId, characterId string) error {
	guild, err := s.AppGuild()
	if err != nil {
		return err
	}
	member, err := s.Session.GuildMember(voterId)
	if err != nil {
		return err
	}
	channel := guild.ChannelByAppId(models.ChannelTownSquare)
	if channel == nil {
		return errors.New("Town square channel not found")
	}
	CallInteraction(s, testlib.TestInteractionOptions{
		Requester: member.User,
		CommandData: discordgo.ApplicationCommandInteractionData{
			Name: lib.ActionVote,
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Name:  lib.ActionOptionVoteUser,
					Value: characterId,
				},
			},
		},
		ChannelId: channel.Id,
	})
	return nil
}
