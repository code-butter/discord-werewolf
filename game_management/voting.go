package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"

	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func canVote(ia *lib.InteractionArgs) error {
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}
	if !guild.DayNight {
		return lib.NewPermissionDeniedError("It's night time. You can vote tomorrow.")
	}
	alive, err := ia.Interaction.RequesterHasRole(lib.RoleAlive)
	if err != nil {
		return err
	}
	if !alive {
		return lib.NewPermissionDeniedError("The dead cannot vote.")
	}
	townSquareChannel := guild.ChannelByAppId(models.ChannelTownSquare)
	if townSquareChannel.Id != ia.Interaction.ChannelId() {
		return lib.NewPermissionDeniedError("You are not in the town square. Vote there.")
	}
	return nil
}

func voteFor(ia *lib.InteractionArgs) error {
	guildId := ia.Interaction.GuildId()
	userId := ia.Interaction.Requester().ID
	voteForId := ia.Interaction.CommandData().GetOption("user").Value.(string)

	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)
	ctx := do.MustInvoke[context.Context](ia.Injector)

	if voteForId == "" {
		_, err := gorm.G[models.GuildVote](gormDB).
			Where("guild_id = ? AND user_id = ?", guildId, userId).
			Delete(ctx)
		if err != nil {
			return err
		}
		msg := fmt.Sprintf("<@%s> has removed their vote.", userId)
		return ia.Interaction.Respond(msg, false)
	}

	// TODO: introduce logic for double voting

	gvdb := gorm.G[models.GuildVote](gormDB)
	_, err := gvdb.
		Where("guild_id = ? AND user_id = ?", guildId, userId).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			vote := models.GuildVote{
				GuildId:     guildId,
				UserId:      userId,
				VotingForId: voteForId,
			}
			return gvdb.Create(ctx, &vote)
		} else {
			return err
		}
	} else {
		rows, err := gvdb.
			Where("guild_id = ? AND user_id = ?", guildId, userId).
			Update(ctx, "voting_for_id", voteForId)
		if rows != 1 {
			return errors.New("Invalid number of votes cast")
		}
		if err != nil {
			return err
		}
	}

	msg := fmt.Sprintf("<@%s> has voted for <@%s>", userId, voteForId)
	return ia.Interaction.Respond(msg, false)
}
