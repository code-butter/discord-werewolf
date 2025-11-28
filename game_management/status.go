package game_management

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"
	"slices"

	"github.com/samber/do"
	"gorm.io/gorm"
)

func whoIsAlive(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.DeferredResponse("Thinking...", false); err != nil {
		return err
	}
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}

	// Report who is alive if game is not started yet
	if !guild.GameGoing {
		playingMembers, err := ia.Session.GuildMembersWithRole(lib.RolePlaying)
		if err != nil {
			return err
		}
		msg := "Playing:"
		for _, member := range playingMembers {
			msg += fmt.Sprintf("\n <@%s>", member.User.ID)
		}
		if len(playingMembers) == 0 {
			msg += "\n (none)"
		}
		return ia.Interaction.FollowupMessage(msg, false)
	}

	// Find out who is alive, and who is dead. #vizinni
	aliveMembers, err := ia.Session.GuildMembersWithRole(lib.RoleAlive)
	if err != nil {
		return err
	}
	characters, err := ia.GuildCharacters()
	if err != nil {
		return err
	}
	aliveIds := make([]string, 0, len(aliveMembers))
	for _, m := range aliveMembers {
		aliveIds = append(aliveIds, m.User.ID)
	}
	aliveCharacters := make([]*models.GuildCharacter, 0, len(aliveMembers))
	deadCharacters := make([]*models.GuildCharacter, 0, len(characters)-len(aliveMembers))
	for _, character := range characters {
		if slices.Contains(aliveIds, character.Id) {
			aliveCharacters = append(aliveCharacters, character)
		} else {
			deadCharacters = append(deadCharacters, character)
		}
	}
	msg := aliveAndDeadList(aliveCharacters, deadCharacters, false, true)
	return ia.Interaction.FollowupMessage(msg, false)
}

func showVotesFor(ia *lib.InteractionArgs) error {
	votersMap, err := getVotersMap(ia)
	if err != nil {
		return err
	}
	count := 0
	msg := ""
	for votedForId, voters := range votersMap {
		msg += fmt.Sprintf("<@%s>: %d\n", votedForId, len(voters))
		count++
	}
	if count == 0 {
		msg += "(none)"
	}
	return ia.Interaction.Respond(msg, false)
}

func showVotersFor(ia *lib.InteractionArgs) error {
	votersMap, err := getVotersMap(ia)
	if err != nil {
		return err
	}
	msg := ""
	count := 0
	for votedForId, voters := range votersMap {
		msg += fmt.Sprintf("<@%s>\n", votedForId)
		for _, voter := range voters {
			msg += fmt.Sprintf(" - <@%s>\n", voter)
		}
		count++
	}
	if count == 0 {
		msg += "(none)"
	}
	return ia.Interaction.Respond(msg, false)
}

func getVotersMap(ia *lib.InteractionArgs) (map[string][]string, error) {
	db := do.MustInvoke[*gorm.DB](ia.Injector)
	ctx := do.MustInvoke[context.Context](ia.Injector)

	votes, err := gorm.G[models.GuildVote](db).Where("guild_id = ?", ia.GuildId).Find(ctx)
	if err != nil {
		return nil, err
	}
	voteMap := make(map[string][]string)
	for _, vote := range votes {
		voteMap[vote.VotingForId] = append(voteMap[vote.VotingForId], vote.UserId)
	}
	return voteMap, nil
}
