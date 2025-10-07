package werewolves

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/listeners"
	"discord-werewolf/lib/models"
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Setup() error {
	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "kill",
			Description: "Vote for a user to kill come day time",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select a player",
					Required:    false,
				},
			}},
		Global:      true,
		Respond:     voteKill,
		Authorizers: []lib.CommandAuthorizer{canKill},
	})

	listeners.GameStartListeners.Add(startGameListener)

	return nil
}

func canKill(args *lib.InteractionArgs) (bool, error) {
	var result *gorm.DB
	var character models.GuildCharacter
	result = args.GormDB.
		Where("guild_id = ? AND user_id = ?", args.Interaction.GuildId(), args.Interaction.Requester().ID).
		First(&character)
	if result.Error != nil {
		return false, result.Error
	}
	if character.CharacterId != models.CharacterWolf && character.CharacterId != models.CharacterWolfCub {
		return false, nil
	}
	return true, nil
}

func voteKill(args *lib.InteractionArgs) error {
	// TODO: figure out how to accurately track double kill ability
	var result *gorm.DB
	var vote *WerewolfKillVote
	guildId := args.Interaction.GuildId()
	requesterId := args.Interaction.Requester().ID
	result = args.GormDB.
		Where("guild_id = ? AND user_id = ?", guildId, requesterId).
		Find(&vote)
	if result.Error != nil {
		return result.Error
	}
	if vote == nil {
		vote = &WerewolfKillVote{
			GuildId:     guildId,
			UserId:      requesterId,
			VotingForId: args.Interaction.CommandData().GetOption("user").Value.(string),
		}
		result = args.GormDB.Create(vote)
	} else {
		result = args.GormDB.Model(&vote).
			Where("guild_id = ? AND user_id = ?").
			Updates(vote)
	}
	if result.Error != nil {
		return result.Error
	}
	msg := fmt.Sprintf("Voted to kill <@%s>", vote.VotingForId)
	return args.Interaction.Respond(msg, false)
}

func startGameListener(s *lib.ServiceArgs, data listeners.GameStartData) error {
	var err error
	wolvesChannel := data.Guild.ChannelByAppId(models.ChannelWerewolves)
	if wolvesChannel == nil {
		return errors.New("could not find wolves channel")
	}
	var postPermissions int64 = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages
	var wolfMentions []string
	for _, character := range data.Characters {
		if character.CharacterId == models.CharacterWolf || character.CharacterId == models.CharacterWolfCub {
			if err = s.Session.UserChannelPermissions(wolvesChannel.Id, character.Id, postPermissions, 0); err != nil {
				return errors.Wrap(err, "could not set post permissions for wolf channel")
			}
			wolfMentions = append(wolfMentions, fmt.Sprintf("<@%x>", character.Id))
		}
	}
	msg := heredoc.Doc(`
		Welcome to the werewolf channel! Talk to your fellow werewolves and mark your next target with the /kill command at night to eat the villagers. Each werewolf can mark their own target. If the werewolf target do not match the werewolf bot will choose the villager to be eaten.
		
		Werewolves:
	`)
	wolfMsg := strings.Join(wolfMentions, ", ")
	return s.Session.Message(wolvesChannel.Id, msg+wolfMsg)

}
