package testlib

import (
	"context"
	"database/sql"
	"discord-werewolf/lib"
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit(session lib.DiscordSession) (lib.SessionArgs, lib.Clock) {
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

	if err := lib.MigrateUp(db); err != nil {
		log.Fatal(err)
	}

	clock := NewMockClock(time.Now())

	injector := do.New()
	do.ProvideValue[*gorm.DB](injector, gormDB)
	do.ProvideValue[context.Context](injector, context.Background())
	do.ProvideValue[lib.Clock](injector, clock)

	return lib.SessionArgs{
		Session:  session,
		Injector: injector,
	}, clock

}

func InteractionInit(session lib.DiscordSession, options TestInteractionOptions) (lib.InteractionArgs, lib.Clock) {
	sa, clock := TestInit(session)
	return lib.InteractionArgs{
		SessionArgs: sa,
		Interaction: NewTestInteraction(session, options),
	}, clock
}
