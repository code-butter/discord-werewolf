package lib

import (
	"context"
	"database/sql"
	"embed"

	"gorm.io/gorm"
)

var DB *sql.DB
var Ctx context.Context
var GormDB *gorm.DB

//go:embed migrations/*.sql
var EmbedMigrations embed.FS
