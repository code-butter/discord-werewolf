package models

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type CharacterAction struct {
	GuildId  string
	UserId   string
	TargetId string
	Action   string
}

func AddOrUpdateCharacterAction(ctx context.Context, db *gorm.DB, action *CharacterAction) error {
	var found = false
	dbCtx := gorm.G[CharacterAction](db)
	_, err := dbCtx.
		Where("guild_id = ? AND user_id = ? AND action = ?", action.GuildId, action.UserId, action.Action).
		First(ctx)
	if err == nil {
		found = true
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Wrap(err, "Could not find existing action")
	}
	if found {
		_, err = dbCtx.
			Where("guild_id = ? AND user_id = ? AND action = ?", action.GuildId, action.UserId, action.Action).
			Updates(ctx, *action)
	} else {
		err = dbCtx.Create(ctx, action)
	}
	if err != nil {
		return errors.Wrap(err, "Failed to save character action")
	}
	return nil
}
