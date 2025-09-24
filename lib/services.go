package lib

import (
	"context"
	"database/sql"

	"gorm.io/gorm"
)

var DB *sql.DB
var Ctx context.Context
var GormDB *gorm.DB
