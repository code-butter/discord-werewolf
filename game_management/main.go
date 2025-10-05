package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
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

	return nil
}

func canVote(ia *lib.InteractionArgs) (bool, error) {
	// TODO: implement me
	return true, nil
}

func voteFor(ia *lib.InteractionArgs) error {
	// TODO: check if vote should happen
	guildId := ia.Interaction.GuildId()
	userId := ia.Interaction.Requester().ID
	voteForId := ia.Interaction.CommandData().GetOption("target").Value.(string)

	if voteForId == "" {
		_, err := gorm.G[models.GuildVote](ia.GormDB).
			Where("guild_id = ? AND user_id = ?", guildId, userId).
			Delete(ia.Ctx)
		return err
	} else {
		vote := models.GuildVote{
			GuildId:     ia.Interaction.GuildId(),
			UserId:      ia.Interaction.Requester().ID,
			VotingForId: voteForId,
		}
		result := ia.GormDB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "guild_id"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "channels"}),
		}).Create(&vote)
		return result.Error
	}
}

// TODO: implement different game modes
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
	players, err := ia.Session.GuildMembersWithRole(lib.RolePlaying)
	if err != nil {
		return err
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

	// TODO: extract this out to wolves' module with a start game listener/callback
	wolvesChannel := guild.ChannelById(models.ChannelWerewolves)
	var postPermissions int64 = discordgo.PermissionViewChannel | discordgo.PermissionSendMessages
	if wolvesChannel == nil {
		return errors.New("could not find wolves channel")
	}
	for _, character := range characters {
		if character.CharacterId == models.CharacterWolf {
			if err = ia.Session.UserChannelPermissions(wolvesChannel.Id, character.Id, postPermissions, 0); err != nil {
				return errors.Wrap(err, "could not set post permissions")
			}
		}
		if err = ia.Session.RemoveRole(character.Id, lib.RolePlaying); err != nil {
			return errors.Wrap(err, "could not remove playing role from user")
		}
		if err = ia.Session.AssignRole(character.Id, lib.RoleAlive); err != nil {
			return errors.Wrap(err, "could not assign alive role to user")
		}
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
