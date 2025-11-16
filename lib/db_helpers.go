package lib

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

func MigrateUp(db *sql.DB) error {
	var err error
	if err = goose.SetDialect("sqlite3"); err != nil {
		return errors.Wrap(err, "Could not set goose dialect")
	}
	// TODO: check if migrations are needed and back up database
	goose.SetBaseFS(EmbedMigrations)

	if err = goose.Up(db, "migrations"); err != nil {
		return errors.Wrap(err, "Could not do auto-migrations")
	}
	return nil
}
