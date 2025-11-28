package testlib

import (
	"context"
	"database/sql"
	"discord-werewolf/lib"
	"discord-werewolf/lib/shared"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/samber/do"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestInitCallback func(injector *do.Injector)

func TestInitDefault(session lib.DiscordSession) *lib.SessionArgs {
	return TestInit(session, func(*do.Injector) {})
}

func TestInit(session lib.DiscordSession, callback TestInitCallback) *lib.SessionArgs {
	var err error

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not create in-memory database"))
	}
	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}))

	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not connect to database with Gorm"))
	}

	goose.SetLogger(NoopLogger{})
	if err := lib.MigrateUp(db); err != nil {
		log.Fatal(err)
	}
	guild, _ := session.Guild()
	sessionProvider := NewTestDiscordSessionProvider(map[string]lib.DiscordSession{guild.ID: session})

	injector := shared.SetupInjector()
	do.ProvideValue[*gorm.DB](injector, gormDB)
	do.ProvideValue[context.Context](injector, context.Background())
	do.ProvideValue[lib.DiscordSessionProvider](injector, sessionProvider)

	callback(injector)

	if _, err := do.Invoke[lib.Clock](injector); err != nil {
		do.ProvideValue[lib.Clock](injector, NewMockClock(time.Now()))
	}

	return &lib.SessionArgs{
		Session:  session,
		Injector: injector,
	}
}

func InteractionInit(args *lib.SessionArgs, options TestInteractionOptions) lib.InteractionArgs {
	return lib.InteractionArgs{
		SessionArgs: args,
		Interaction: NewTestInteraction(args, options),
	}
}

func GenericServerInit(memberCount int, callback TestInitCallback) *lib.SessionArgs {
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
	session := NewTestSession(TestSessionOptions{
		GuildRoles: roles,
		Owner:      owner,
	})
	sessionArgs := TestInit(session, callback)
	var members []*discordgo.Member
	for i := 0; i < memberCount; i++ {
		members = append(members, TestDiscordMember(session.GuildId))
	}
	session.Members = append(members, &discordgo.Member{
		GuildID:  session.GuildId,
		JoinedAt: time.Now().UTC().Add(-time.Hour * 365),
		Nick:     "da boss",
		User:     owner,
	})
	args := InteractionInit(sessionArgs, TestInteractionOptions{
		Requester: owner,
	})
	if err := shared.InitGuild(&args); err != nil {
		log.Fatal(err)
	}
	return sessionArgs
}

func StartTestGameDefault(memberCount int, playingCount int) *lib.SessionArgs {
	return StartTestGame(memberCount, playingCount, func(injector *do.Injector) {})
}

func StartTestGame(memberCount int, playingCount int, callback TestInitCallback) *lib.SessionArgs {
	args := GenericServerInit(memberCount, callback)
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
	ownerInteraction := InteractionInit(args, TestInteractionOptions{
		Requester: owner,
	})
	if err := shared.StartGame(&ownerInteraction); err != nil {
		log.Fatal(err)
	}
	return args
}
