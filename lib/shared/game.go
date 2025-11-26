package shared

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"math/rand"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func StartGame(ia *lib.InteractionArgs) error {
	var err error
	var result *gorm.DB
	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)
	listeners := do.MustInvoke[*lib.GameListeners](ia.Injector)
	settings := do.MustInvoke[*lib.GuildSettings](ia.Injector)
	ctx := do.MustInvoke[context.Context](ia.Injector)

	if err = ia.Interaction.DeferredResponse("Starting game...", true); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = ia.Interaction.FollowupMessage("Server error starting game.", true)
		}
	}()

	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}
	if guild.GameGoing {
		return ia.Interaction.FollowupMessage("Game already started.", true)
	}

	_, err = gorm.G[models.Guild](gormDB).
		Where("id = ?", guild.Id).
		Update(ctx, "game_going", 1)

	if err != nil {
		return err
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

	if result = gormDB.Save(characters); result.Error != nil {
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

	if err = settings.StartGame(guild.Id); err != nil {
		return err
	}

	err = listeners.GameStart.Trigger(ia.SessionArgs, lib.GameStartData{
		Guild:      guild,
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

func removeAllFromRole(s lib.DiscordSession, role string) error {
	members, err := s.GuildMembersWithRole(role)
	if err != nil {
		return err
	}
	for _, member := range members {
		if err := s.RemoveRole(member.User.ID, role); err != nil {
			return err
		}
	}
	return nil
}

func EndGameInteraction(ia *lib.InteractionArgs) error {
	var err error
	if err = ia.Interaction.DeferredResponse("Ending game...", true); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = ia.Interaction.FollowupMessage("Server error ending game.", true)
		}
	}()
	err = EndGame(ia.SessionArgs, discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Title:       "Game Ended",
		Description: "Game ended manually by admin.",
	})
	if err != nil {
		return err
	}
	return ia.Interaction.FollowupMessage("Game ended", true)
}

func EndGame(args *lib.SessionArgs, msg discordgo.MessageEmbed) error {
	var err error
	gormDB := do.MustInvoke[*gorm.DB](args.Injector)
	ctx := do.MustInvoke[context.Context](args.Injector)
	l := do.MustInvoke[*lib.GameListeners](args.Injector)
	guild, err := args.AppGuild()
	if err != nil {
		return err
	}
	if err = removeAllFromRole(args.Session, lib.RoleAlive); err != nil {
		return err
	}
	if err = removeAllFromRole(args.Session, lib.RoleDead); err != nil {
		return err
	}
	_, err = gorm.G[models.Guild](gormDB).
		Where("id = ?", guild.Id).
		Update(ctx, "game_going", 0)
	if err != nil {
		return err
	}

	err = l.GameEnd.Trigger(args, lib.GameEndData{
		Guild: guild,
	})
	if err != nil {
		return err
	}

	townHall := guild.ChannelByAppId(models.ChannelTownSquare)
	if townHall == nil {
		return errors.New("could not find town square channel")
	}
	return args.Session.MessageEmbed(townHall.Id, &msg)
}

func StartDay(s *lib.SessionArgs) error {
	var err error

	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	l := do.MustInvoke[*lib.GameListeners](s.Injector)

	guild, err := s.AppGuild()
	if err != nil {
		return err
	}

	err = l.DayStart.Trigger(s, lib.DayStartData{
		Guild: guild,
	})
	if err != nil {
		if gameOver, ok := err.(lib.GameOver); ok {
			return EndGame(s, gameOver.MessageEmbed)
		}
		return errors.Wrap(err, "Failed to start day from triggers")
	}

	guild.DayNight = true
	if result := gormDB.Save(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Could not save guild")
	}

	return nil
}

func StartNight(s *lib.SessionArgs) error {
	var err error

	gormDB := do.MustInvoke[*gorm.DB](s.Injector)
	l := do.MustInvoke[*lib.GameListeners](s.Injector)

	guild, err := s.AppGuild()
	if err != nil {
		return err
	}
	err = l.NightStart.Trigger(s, lib.NightStartData{
		Guild: guild,
	})
	if err != nil {
		if gameOver, ok := err.(lib.GameOver); ok {
			return EndGame(s, gameOver.MessageEmbed)
		}
		return errors.Wrap(err, "Failed to start night from triggers")
	}
	guild.DayNight = false
	if result := gormDB.Save(&guild); result.Error != nil {
		return errors.Wrap(result.Error, "Could not save guild")
	}
	return nil
}

func KillCharacter(s *lib.SessionArgs, character *models.GuildCharacter, cause string) error {
	var err error

	db := do.MustInvoke[*gorm.DB](s.Injector)
	ctx := do.MustInvoke[context.Context](s.Injector)
	l := do.MustInvoke[*lib.GameListeners](s.Injector)

	guild, err := s.AppGuild()
	if err != nil {
		return err
	}
	if character == nil {
		return errors.New("character to kill is nil")
	}

	if err = s.Session.RemoveRole(character.Id, "Alive"); err != nil {
		return errors.Wrap(err, "Could not remove Alive role")
	}
	if err = s.Session.AssignRole(character.Id, "Dead"); err != nil {
		return errors.Wrap(err, "Could not assign Dead role")
	}

	character.ExtraData["death_cause"] = cause
	_, err = gorm.G[models.GuildCharacter](db).
		Updates(ctx, *character)
	if err != nil {
		return errors.Wrap(err, "could not update character with ID "+character.Id)
	}
	return l.CharacterDeath.Trigger(s, lib.CharacterDeathData{
		Guild:  guild,
		Target: *character,
	})
}
