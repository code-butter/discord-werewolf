package lib

import (
	"encoding/json"
	"fmt"
	"log"

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

func MigrateUp() {
	var err error
	if err = goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(errors.Wrap(err, "Could not set goose dialect"))
	}

	// TODO: check if migrations are needed and back up database
	goose.SetBaseFS(EmbedMigrations)
	if err = goose.Up(DB, "migrations"); err != nil {
		log.Fatal(errors.Wrap(err, "Could not do auto-migrations"))
	}
}
