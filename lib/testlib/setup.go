package testlib

import (
	"context"
	"database/sql"
	"discord-werewolf/lib"
	"log"

	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInit() {
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

	lib.DB = db
	lib.Ctx = context.Background() // TODO: make this listen to signals
	lib.GormDB = gormDB

	lib.MigrateUp()

}
