package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"

	"github.com/samber/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func canVote(ia *lib.InteractionArgs) error {
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}
	wolfChannel := guild.ChannelByAppId(models.ChannelTownSquare)
	if wolfChannel.Id != ia.Interaction.ChannelId() {
		return lib.NewPermissionDeniedError("You are not in the town square. Vote there.")
	}
	return nil
}

func voteFor(ia *lib.InteractionArgs) error {
	// TODO: check if vote should happen
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

	vote := models.GuildVote{
		GuildId:     ia.Interaction.GuildId(),
		UserId:      ia.Interaction.Requester().ID,
		VotingForId: voteForId,
	}
	result := gormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"voting_for_id"}),
	}).Create(&vote)
	if result.Error != nil {
		return result.Error
	}
	msg := fmt.Sprintf("<@%s> has voted for <@%s>", userId, voteForId)
	return ia.Interaction.Respond(msg, false)
}
