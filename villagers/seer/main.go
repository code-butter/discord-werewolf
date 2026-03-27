package main

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/authorizors"
	"discord-werewolf/lib/characters"
	"discord-werewolf/lib/models"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func Setup(injector *do.Injector) (err error) {
	l := do.MustInvoke[*lib.GameListeners](injector)
	cr := do.MustInvoke[*lib.CommandRegistry](injector)

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionInvestigate,
			Description: "Investigate a player (seer)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        lib.ActionOptionInvestigateUser,
					Description: "Select a player",
					Required:    true,
				},
			},
		},
		Respond: investigatePlayer,
		Authorizers: []lib.Authorizer{
			authorizors.CharacterExists(lib.ActionOptionInvestigateUser),
			canInvestigate,
			authorizors.IsAlive,
			authorizors.IsNightTime,
		},
	})

	l.GameStart.Add(startGameListener)
	l.DayStart.Add(dayStartListener)

	return nil
}

func dayStartListener(s *lib.SessionArgs, data lib.DayStartData) error {
	return nil
}

func startGameListener(s *lib.SessionArgs, data lib.GameStartData) error {
	return nil
}

func investigatePlayer(ia *lib.InteractionArgs) error {
	db := do.MustInvoke[*gorm.DB](ia.Injector)
	ctx := do.MustInvoke[context.Context](ia.Injector)

	err := models.AddOrUpdateCharacterAction(ctx, db, &models.CharacterAction{
		GuildId:  ia.Interaction.GuildId(),
		UserId:   ia.Interaction.Requester().ID,
		Action:   lib.ActionInvestigate,
		TargetId: ia.Interaction.CommandData().GetOption(lib.ActionOptionInvestigateUser).Value.(string),
	})

	if err != nil {
		return err
	}

	// TODO: send confirmation message

	return nil
}

func canInvestigate(ia *lib.InteractionArgs) (err error) {
	character, err := ia.GuildCharacter(ia.Interaction.Requester().ID)
	if err != nil {
		return
	}
	if character.CharacterId != characters.Seer && character.CharacterId != characters.Fool {
		return lib.NewPermissionDeniedError("You're not a seer!")
	}
	//guild, err := ia.AppGuild()
	//if err != nil {
	//	return
	//}
	// TODO: check if seer channel and player is assigned

	return nil
}
