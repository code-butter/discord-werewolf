package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/listeners"
	"discord-werewolf/lib/models"
	"fmt"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Setup() error {
	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "playing",
			Description: "Signs you up for the next round.",
		},
		Global:  true,
		Respond: playing,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "stop_playing",
			Description: "Removes you from playing next round.",
		},
		Global:  true,
		Respond: stopPlaying,
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "start_game",
			Description: "Starts the game.",
		},
		Global:      true,
		Respond:     startGame,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "day_time",
			Description: "Triggers day for the current game",
		},
		Global:      true,
		Respond:     triggerDay,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "night_time",
			Description: "Triggers night for the current game",
		},
		Global:      true,
		Respond:     triggerNight,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "vote",
			Description: "Vote to hang. Leave off the target if you wish to unvote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "Select a player",
					Required:    false,
				},
			},
		},
		Global:      true,
		Respond:     voteFor,
		Authorizers: []lib.CommandAuthorizer{canVote},
	})

	listeners.NightStartListeners.Add(hangCharacters)

	return nil
}

func triggerNight(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering night...", true)
	if err != nil {
		return err
	}
	if err = StartNight(args.Interaction.GuildId(), *args.ServiceArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Night triggered!", true)
}

func triggerDay(args *lib.InteractionArgs) error {
	err := args.Interaction.DeferredResponse("Triggering day...", true)
	if err != nil {
		return err
	}
	if err = StartDay(args.Interaction.GuildId(), *args.ServiceArgs); err != nil {
		return err
	}
	return args.Interaction.FollowupMessage("Day triggered!", true)
}

func hangCharacters(s *lib.ServiceArgs, data listeners.NightStartData) error {
	var err error
	var result *gorm.DB
	townChannel := data.Guild.ChannelByAppId(models.ChannelTownSquare)
	if townChannel == nil {
		return errors.New("No town channel found for guild " + data.Guild.Id)
	}
	var voted string
	result = s.GormDB.
		Model(&models.GuildVote{}).
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
			Title:       "No one voted.",
			Description: "No one was hanged.",
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
	} else {
		var character models.GuildCharacter
		result = s.GormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).First(&character)
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
			Type:  discordgo.EmbedTypeRich,
			Title: fmt.Sprintf("The town has hanged <@%s>.", voted),
		}
		if err = s.Session.MessageEmbed(townChannel.Id, msg); err != nil {
			return errors.Wrap(err, "Could not send town channel message")
		}
		character.ExtraData["death_cause"] = "hanged"
		if result = s.GormDB.Where("id = ? AND guild_id = ?", voted, data.Guild.Id).Save(&character); result.Error != nil {
			return errors.Wrap(result.Error, "Could not update character with ID "+voted)
		}
	}
	_, err = gorm.G[models.GuildVote](s.GormDB).Where("guild_id = ?", data.Guild.Id).Delete(s.Ctx)
	return err
}

func canVote(ia *lib.InteractionArgs) (bool, error) {
	// TODO: implement me
	return true, nil
}

func voteFor(ia *lib.InteractionArgs) error {
	// TODO: check if vote should happen
	guildId := ia.Interaction.GuildId()
	userId := ia.Interaction.Requester().ID
	voteForId := ia.Interaction.CommandData().GetOption("user").Value.(string)

	if voteForId == "" {
		_, err := gorm.G[models.GuildVote](ia.GormDB).
			Where("guild_id = ? AND user_id = ?", guildId, userId).
			Delete(ia.Ctx)
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
	result := ia.GormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "guild_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"voting_for_id"}),
	}).Create(&vote)
	if result.Error != nil {
		return result.Error
	}
	msg := fmt.Sprintf("<@%s> has voted for <@%s>", userId, voteForId)
	return ia.Interaction.Respond(msg, false)
}

// TODO: implement different game modes
// TODO: enable scheduled start
func startGame(ia *lib.InteractionArgs) error {
	var err error
	if err = ia.Interaction.DeferredResponse("Starting game...", true); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = ia.Interaction.FollowupMessage("Server error starting game.", true)
		}
	}()

	var guild *models.Guild
	var result *gorm.DB
	if result = ia.GormDB.Where("id = ?", ia.Interaction.GuildId()).First(&guild); result.Error != nil {
		return result.Error
	}
	if result = ia.GormDB.Where("guild_id = ?", ia.Interaction.GuildId()).Delete(&models.GuildCharacter{}); result.Error != nil {
		return result.Error
	}
	if result = ia.GormDB.Where("guild_id = ?", ia.Interaction.GuildId()).Delete(&models.GuildVote{}); result.Error != nil {
		return result.Error
	}

	// Clear town channels
	catTown := guild.ChannelByAppId(models.CatChannelTheTown)
	for _, channel := range *catTown.Children {
		if err = ia.Session.ClearChannelMessages(channel.Id); err != nil {
			return err
		}
	}

	players, err := ia.Session.GuildMembersWithRole(lib.RolePlaying)
	if err != nil {
		return err
	}

	if len(players) == 0 {
		return ia.Interaction.FollowupMessage("Nobody is playing!", true)
	}

	// Randomly mix players
	for i := range players {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	characters := make([]models.GuildCharacter, 0)

	for i, player := range players {
		character := models.GuildCharacter{
			Id:      player.User.ID,
			GuildId: ia.Interaction.GuildId(),
		}
		if i%3 == 0 {
			character.CharacterId = models.CharacterWolf
		} else {
			character.CharacterId = models.CharacterVillager
		}
		characters = append(characters, character)
	}

	if result = ia.GormDB.Save(&characters); result.Error != nil {
		return result.Error
	}

	for _, character := range characters {
		if err = ia.Session.RemoveRole(character.Id, lib.RolePlaying); err != nil {
			return errors.Wrap(err, "could not remove playing role from user")
		}
		if err = ia.Session.AssignRole(character.Id, lib.RoleAlive); err != nil {
			return errors.Wrap(err, "could not assign alive role to user")
		}
	}

	err = listeners.GameStartListeners.Trigger(ia.ServiceArgs, listeners.GameStartData{
		Guild:      *guild,
		Characters: characters,
	})
	if err != nil {
		return err
	}

	townChannel := guild.ChannelByAppId(models.ChannelTownSquare)
	if townChannel == nil {
		return errors.New("could not find town square channel")
	}

	err = ia.Session.Message(townChannel.Id, "Welcome to the town-square! Here you will vote for who you think the werewolves are.")
	if err != nil {
		return errors.Wrap(err, "could not send welcome message to town square")
	}

	err = ia.Interaction.FollowupMessage("Game started", true)
	if err != nil {
		return errors.Wrap(err, "could not follow up message")
	}
	return nil
}

func playing(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.AssignRoleToRequester(lib.RolePlaying); err != nil {
		return err
	}
	return ia.Interaction.Respond("Now playing!", false)
}

func stopPlaying(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.RemoveRoleFromRequester(lib.RolePlaying); err != nil {
		return err
	}
	return ia.Interaction.Respond("Stopped playing.", false)
}
