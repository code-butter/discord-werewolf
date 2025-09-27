package lib

import (
	"context"
	"database/sql"
	"embed"

	"gorm.io/gorm"
)

// TODO: make the databases part of a services struct instead of global for parallel testing

var DB *sql.DB
var Ctx context.Context
var GormDB *gorm.DB

//go:embed migrations/*.sql
var EmbedMigrations embed.FS
