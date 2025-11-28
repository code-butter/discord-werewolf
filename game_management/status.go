package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"
	"slices"
)

func whoIsAlive(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.DeferredResponse("Thinking...", false); err != nil {
		return err
	}
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}

	if guild.GameGoing {
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
