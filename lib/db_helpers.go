package lib

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

func UnMarshalBytes[T any](m *T, value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to unmarshal JSONB from %v", value)
	}
	return json.Unmarshal(bytes, m)
}

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
