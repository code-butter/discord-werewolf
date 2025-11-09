package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/testlib"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func GenericServerInit(memberCount int) lib.SessionArgs {
	owner := &discordgo.User{
		ID: "owner",
	}
	roles := make([]*discordgo.Role, 0)
	for _, role := range lib.Roles {
		roles = append(roles, &discordgo.Role{
			ID:          uuid.NewString(),
			Name:        role.Name,
			Managed:     false,
			Mentionable: true,
			Color:       role.Color,
		})
	}
	session := testlib.NewTestSession(testlib.TestSessionOptions{
		GuildRoles: roles,
		Owner:      owner,
	})
	sessionArgs := testlib.TestInit(session)
	var members []*discordgo.Member
	for i := 0; i < memberCount; i++ {
		members = append(members, testlib.TestDiscordMember(session.GuildId))
	}
	session.Members = append(members, &discordgo.Member{
		GuildID:  session.GuildId,
		JoinedAt: time.Now().UTC().Add(-time.Hour * 365),
		Nick:     "da boss",
		User:     owner,
	})
	args := testlib.InteractionInit(sessionArgs, testlib.TestInteractionOptions{
		Requester: owner,
	})
	if err := InitGuild(&args); err != nil {
		log.Fatal(err)
	}
	return sessionArgs
}

func StartTestGame(memberCount int, playingCount int) lib.SessionArgs {
	args := GenericServerInit(memberCount)
	members, _ := args.Session.GuildMembers()
	guild, _ := args.Session.Guild()
	var owner *discordgo.User
	playingRole, err := args.Session.GetRoleByName(lib.RolePlaying)
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < playingCount; i++ {
		members[i].Roles = append(members[i].Roles, playingRole.ID)
	}
	for _, member := range members {
		if member.User.ID == guild.OwnerID {
			owner = member.User
			break
		}
	}
	ownerInteraction := testlib.InteractionInit(args, testlib.TestInteractionOptions{
		Requester: owner,
	})
	if err := StartGame(&ownerInteraction); err != nil {
		log.Fatal(err)
	}
	return args
}

func TestStartGame(t *testing.T) {
	var result *gorm.DB
	memberCount := rand.Intn(15) + 10
	playingCount := rand.Intn(5) + 5
	args := StartTestGame(memberCount, playingCount)
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
