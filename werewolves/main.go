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
	"github.com/samber/do"
	"gorm.io/gorm"
)

func Setup() error {
	lib.RegisterGlobalCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "kill",
			Description: "Vote for a user to kill over night",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select a player",
					Required:    false,
				},
			}},

		Respond:     voteKill,
		Authorizers: []lib.CommandAuthorizer{canKill},
	})

	listeners.GameStartListeners.Add(startGameListener)
	listeners.DayStartListeners.Add(killVillagers)

	return nil
}

func canKill(ia *lib.InteractionArgs) (bool, error) {
	var result *gorm.DB
	var character models.GuildCharacter

	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)
	result = gormDB.
		Where("guild_id = ? AND id = ?", ia.Interaction.GuildId(), ia.Interaction.Requester().ID).
		First(&character)
	if result.Error != nil {
		return false, result.Error
	}
	if character.CharacterId != models.CharacterWolf && character.CharacterId != models.CharacterWolfCub {
		return false, nil
	}
	return true, nil
}

func voteKill(ia *lib.InteractionArgs) error {
	// TODO: figure out how to accurately track double kill ability
	var result *gorm.DB
	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)
	var vote *WerewolfKillVote
	guildId := ia.Interaction.GuildId()
	requesterId := ia.Interaction.Requester().ID
	result = gormDB.
		Where("guild_id = ? AND user_id = ?", guildId, requesterId).
		First(&vote)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			vote = nil
		} else {
			return errors.Wrap(result.Error, "failed to find vote")
		}
	}
	if vote == nil {
		vote = &WerewolfKillVote{
			GuildId:     guildId,
			UserId:      requesterId,
			VotingForId: ia.Interaction.CommandData().GetOption("user").Value.(string),
		}
		result = gormDB.Create(vote)
	} else {
		vote.VotingForId = ia.Interaction.CommandData().GetOption("user").Value.(string)
		result = gormDB.Model(&vote).
			Where("guild_id = ? AND user_id = ?", guildId, requesterId).
			Updates(vote)
	}
	if result.Error != nil {
		return errors.Wrap(result.Error, "failed to save vote")
	}
	msg := fmt.Sprintf("Voted to kill <@%s>", vote.VotingForId)
	return ia.Interaction.Respond(msg, false)
}

func startGameListener(s *lib.SessionArgs, data listeners.GameStartData) error {
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
			wolfMentions = append(wolfMentions, fmt.Sprintf("<@%s>", character.Id))
		}
	}
	msg := heredoc.Doc(`
		Welcome to the werewolf channel! Talk to your fellow werewolves and mark your next target with the ` + "`/kill`" + ` command at night to eat the villagers. Each werewolf can mark their own target. If the werewolf target do not match the werewolf bot will choose the villager to be eaten.
		
		Werewolves:
	`)
	wolfMsg := strings.Join(wolfMentions, ", ")
	return s.Session.Message(wolvesChannel.Id, msg+wolfMsg)
}
func killVillagers(s *lib.SessionArgs, data listeners.DayStartData) error {
	var err error
	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	var result *gorm.DB
	townChannel := data.Guild.ChannelByAppId(models.ChannelTownSquare)
	if townChannel == nil {
		return errors.New("No town channel found for guild " + data.Guild.Id)
	}
	var voted string
	result = gormDB.
		Model(&WerewolfKillVote{}).
		Select("voting_for_id").
		Group("voting_for_id").
		Order("COUNT(*) DESC").
		Where("guild_id = ?", data.Guild.Id).
		Limit(1).
		Pluck("voting_for_id", &voted)
	if result.Error != nil {
		msg := fmt.Sprintf("Could not get votes for guild %s with ID %s", data.Guild.Name, data.Guild.Id)
		return errors.Wrap(result.Error, msg)
	}

	if voted == "" {
		msg := &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "No one died last night.",
			Description: "What a bunch of lazy wolves.",
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
	} else {
		var character models.GuildCharacter
		result = gormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).First(&character)
		if result.Error != nil {
			return errors.Wrap(result.Error, "Could not find character with ID "+voted)
		}
		if err = s.Session.RemoveRole(voted, "Alive"); err != nil {
			return errors.Wrap(err, "Could not remove Alive role")
		}
		if err = s.Session.AssignRole(voted, "Dead"); err != nil {
			return errors.Wrap(err, "Could not assign Dead role")
		}
		msg := &discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Title:       "Last night the werwolves attacked!",
			Description: fmt.Sprintf("<@%s> has died at the fangs of wolves.", voted),
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
		character.ExtraData["death_cause"] = "wolf"
		if result = gormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).Save(&character); result.Error != nil {
			return errors.Wrap(result.Error, "Could not update character with ID "+voted)
		}
	}
	return nil
}
