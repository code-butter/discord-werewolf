package lib

import (
	"context"

	"gorm.io/gorm"
)

type ServiceArgs struct {
	Session DiscordSession
	GormDB  *gorm.DB
	Ctx     context.Context
}
