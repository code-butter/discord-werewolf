package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/listeners"
	"discord-werewolf/lib/models"
	"math/rand"

	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

// TODO: implement different game modes
// TODO: enable scheduled start
// This is public for tests in other packages
func StartGame(ia *lib.InteractionArgs) error {
	var err error
	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)
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
	if result = gormDB.Where("id = ?", ia.Interaction.GuildId()).First(&guild); result.Error != nil {
		return result.Error
	}
	if result = gormDB.Where("guild_id = ?", ia.Interaction.GuildId()).Delete(&models.GuildCharacter{}); result.Error != nil {
		return result.Error
	}
	if result = gormDB.Where("guild_id = ?", ia.Interaction.GuildId()).Delete(&models.GuildVote{}); result.Error != nil {
		return result.Error
	}

	// Check if we can start
	players, err := ia.Session.GuildMembersWithRole(lib.RolePlaying)
	if err != nil {
		return err
	}

	if len(players) == 0 {
		return ia.Interaction.FollowupMessage("Nobody is playing!", true)
	}

	// Clear town channels
	catTown := guild.ChannelByAppId(models.CatChannelTheTown)
	for _, channel := range *catTown.Children {
		if err = ia.Session.ClearChannelMessages(channel.Id); err != nil {
			return err
		}
	}

	// Randomly mix players
	for i := range players {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}

	characters := make([]models.GuildCharacter, 0)

	for i, player := range players {
		character := models.GuildCharacter{
			Id:        player.User.ID,
			GuildId:   ia.Interaction.GuildId(),
			ExtraData: models.JsonMap{},
		}
		if i%5 == 0 {
			character.CharacterId = models.CharacterWolf
		} else {
			character.CharacterId = models.CharacterVillager
		}
		characters = append(characters, character)
	}

	if result = gormDB.Save(&characters); result.Error != nil {
		return result.Error
	}

	for _, character := range characters {
		if err = ia.Session.RemoveRole(character.Id, lib.RolePlaying); err != nil {
			return errors.Wrap(err, "could not remove playing role from user")
		}
		if err = ia.Session.RemoveRole(character.Id, lib.RoleDead); err != nil {
			return errors.Wrap(err, "could not remove dead role from user")
		}
		if err = ia.Session.AssignRole(character.Id, lib.RoleAlive); err != nil {
			return errors.Wrap(err, "could not assign alive role to user")
		}
	}

	guild.GameGoing = true
	if result = gormDB.Save(&guild); result.Error != nil {
		return result.Error
	}

	err = listeners.GameStartListeners.Trigger(&ia.SessionArgs, listeners.GameStartData{
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
