package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/shared"
	"discord-werewolf/lib/testlib"
	"math/rand"
	"testing"

	"github.com/samber/do"
	"gorm.io/gorm"
)

func TestStartGame(t *testing.T) {
	var result *gorm.DB
	memberCount := rand.Intn(15) + 10
	playingCount := rand.Intn(5) + 5
	args := testlib.StartTestGameDefault(memberCount, playingCount)
	guild, _ := args.Session.Guild()
	gormDb := do.MustInvoke[*gorm.DB](args.Injector)
	membersAlive, _ := args.Session.GuildMembersWithRole(lib.RoleAlive)
	aliveCount := len(membersAlive)

	if playingCount != aliveCount {
		t.Errorf("alive count should be %d got %d", playingCount, aliveCount)
	}

	var gameGoing int
	result = gormDb.
		Model(&models.Guild{}).
		Select("game_going").
		Where("id = ?", guild.ID).
		Pluck("game_going", &gameGoing)

	if gameGoing != 1 {
		t.Errorf("game did not start")
	}

	// TODO: break this out into different game mode tests
	var actualWolfCount int
	result = gormDb.
		Model(&models.GuildCharacter{}).
		Select("count(*) as cnt").
		Where("character_id = ? AND guild_id = ?", models.CharacterWolf, guild.ID).
		Pluck("cnt", &actualWolfCount)
	if result.Error != nil {
		t.Error(result.Error)
	}
	wolfCount := playingCount / 5
	if playingCount%5 != 0 {
		wolfCount++
	}
	if wolfCount != actualWolfCount {
		t.Errorf("wolf count should be %d got %d", wolfCount, actualWolfCount)
	}

}

func TestEndGame(t *testing.T) {
	args := testlib.StartTestGameDefault(10, 10)

	interaction := testlib.NewTestInteraction(args, testlib.TestInteractionOptions{})
	interactionArgs := &lib.InteractionArgs{
		SessionArgs: args,
		Interaction: interaction,
	}
	if err := shared.EndGame(interactionArgs); err != nil {
		t.Fatal(err)
	}

	guild, _ := args.AppGuild()
	if guild.GameGoing {
		t.Errorf("game did not stop")
	}

	membersAlive, _ := args.Session.GuildMembersWithRole(lib.RoleAlive)
	if len(membersAlive) > 0 {
		t.Errorf("alive count should be 0 got %d", len(membersAlive))
	}
	membersDead, _ := args.Session.GuildMembersWithRole(lib.RoleDead)
	if len(membersDead) > 0 {
		t.Errorf("dead count should be 0 got %d", len(membersDead))
	}
}
