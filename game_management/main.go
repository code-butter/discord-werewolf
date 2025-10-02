package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
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
			Description: "Starts the game server.",
		},
		Global:      true,
		Respond:     startGame,
		Authorizers: []lib.CommandAuthorizer{lib.IsAdmin},
	})

	lib.RegisterCommand(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "vote",
			Description: "Vote to hang.",
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

func canVote(i lib.Interaction) (bool, error) {
	// TODO: implement me
	return true, nil
}

func voteFor(i lib.Interaction) error {

}

// TODO: implement different game modes
func startGame(int lib.Interaction) error {
	var err error
	if err = int.DeferredResponse("Starting game...", true); err != nil {
		return err
	}
	var guild *models.Guild
	var result *gorm.DB
	if result = lib.GormDB.Where("id = ?", int.GuildId()).First(&guild); result.Error != nil {
		return result.Error
	}
	if result = lib.GormDB.Where("guild_id = ?", int.GuildId()).Delete(&models.GuildCharacter{}); result.Error != nil {
		return result.Error
	}
	if result = lib.GormDB.Where("guild_id = ?", int.GuildId()).Delete(&models.GuildVote{}); result.Error != nil {
		return result.Error
	}
	players, err := int.GuildMembersWithRole(lib.RolePlaying)
	if err != nil {
		return err
	}

	// Randomly mix players
	for i := range players {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	characters := make([]models.GuildCharacter, len(players))

	for i, player := range players {
		character := models.GuildCharacter{
			Id:      player.User.ID,
			GuildId: int.GuildId(),
		}
		if i%3 == 0 {
			character.CharacterId = models.CharacterWolf
		} else {
			character.CharacterId = models.CharacterVillager
		}
	}

	if result = lib.GormDB.Save(&characters); result.Error != nil {
		return result.Error
	}

	return nil
}

func playing(i lib.Interaction) error {
	if err := i.AssignRoleToRequester(lib.RolePlaying); err != nil {
		return err
	}
	return i.Respond("Now playing!", false)
}

func stopPlaying(i lib.Interaction) error {
	if err := i.RemoveRoleFromRequester(lib.RolePlaying); err != nil {
		return err
	}
	return i.Respond("Stopped playing.", false)
}
