package authorizors

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/testlib"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func TestCharacterExists(t *testing.T) {
	guildId := "guild-id"
	characterId := "character-id"
	aliveRoleId := "alive-role-id"
	member := &discordgo.Member{
		Roles: []string{aliveRoleId},
		User: &discordgo.User{
			ID: characterId,
		},
	}
	session := testlib.NewTestSession(testlib.TestSessionOptions{
		Id:           &guildId,
		GuildMembers: []*discordgo.Member{member},
		GuildRoles: []*discordgo.Role{
			{ID: aliveRoleId, Name: lib.RoleAlive},
		},
		Owner: nil,
	})
	args := testlib.TestInitDefault(session)
	db := do.MustInvoke[*gorm.DB](args.Injector)
	ctx := do.MustInvoke[context.Context](args.Injector)
	err := gorm.G[models.GuildCharacter](db).Create(ctx, &models.GuildCharacter{
		Id:      characterId,
		GuildId: guildId,
	})
	if err != nil {
		t.Fatal(err)
	}
	option := discordgo.ApplicationCommandInteractionDataOption{
		Name:  "target",
		Value: characterId,
	}
	interaction := testlib.NewTestInteraction(args, testlib.TestInteractionOptions{
		CommandData: discordgo.ApplicationCommandInteractionData{
			Options: []*discordgo.ApplicationCommandInteractionDataOption{
				&option,
			},
		},
	})

	// Happy path
	f := CharacterExists("target")
	interactionArgs := &lib.InteractionArgs{
		SessionArgs: args,
		Interaction: interaction,
	}
	err = f(interactionArgs)
	if err != nil {
		t.Fatal(err)
	}

	// Check if player is not in database
	option.Value = "unknown-id"
	err = f(interactionArgs)
	if err == nil {
		t.Error("expected PermissionDeniedError")
	} else if _, ok := err.(lib.PermissionDeniedError); !ok {
		t.Error("expected PermissionDeniedError")
	}

	// Check if target has died
	member.Roles = []string{lib.RoleDead}
	err = f(interactionArgs)
	if err == nil {
		t.Error("expected PermissionDeniedError")
	} else if _, ok := err.(lib.PermissionDeniedError); !ok {
		t.Error("expected PermissionDeniedError")
	}
}
