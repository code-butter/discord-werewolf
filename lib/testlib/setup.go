package testlib

import (
	"context"
	"database/sql"
	"discord-werewolf/lib"
	"log"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit() lib.SessionArgs {
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
	clock.Unfreeze()

	return lib.SessionArgs{
		Session: nil,
		GormDB:  gormDB,
		Ctx:     context.Background(), // TODO: make this listen to signals?
		Clock:   clock,
	}

}
