package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/authorizors"
	"discord-werewolf/lib/models"
	"discord-werewolf/lib/shared"
	"regexp"

	"github.com/bwmarrin/discordgo"
	sets "github.com/hashicorp/go-set/v3"
	"github.com/samber/do"
)

func Setup(injector *do.Injector) error {
	cr := do.MustInvoke[*lib.CommandRegistrar](injector)
	l := do.MustInvoke[*lib.GameListeners](injector)

	l.NightStart.Add(nightListener)
	l.CharacterDeath.Add(checkWinConditions)

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionInit,
			Description: "Initializes the server. Wipes out any data previously stored.",
		},
		Respond:     shared.InitGuild,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Pings the server. Responds with 'pong'.",
		},
		Respond: ping,
	})

	tzs := lib.AllTimeZoneNames()
	locationMatcher := regexp.MustCompile(`^([^/])+`)
	locationSet := sets.New[string](0)
	for _, tz := range tzs {
		area := locationMatcher.FindString(tz)
		if area != "" {
			locationSet.Insert(area)
		}
	}
	var locationChoices []*discordgo.ApplicationCommandOptionChoice
	for tzLocation := range locationSet.Items() {
		if tzLocation == "Etc" {
			continue
		}
		locationChoices = append(locationChoices, &discordgo.ApplicationCommandOptionChoice{
			Name:  tzLocation,
			Value: tzLocation,
		})
	}

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "get_timezones",
			Description: "Get timezones for the server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "area",
					Description: "Get timezones in this general area.",
					Required:    true,
					Choices:     locationChoices,
				},
			},
		},

		Respond:     getTimeZones,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "set_timezone",
			Description: "Sets the timezone for the server.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timezone",
					Description: "Sets the timezone for the server.",
					Required:    true,
				},
			},
		},

		Respond:     setTimeZone,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionPlaying,
			Description: "Signs you up for the next round.",
		},
		Respond: playing,
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionStopPlaying,
			Description: "Removes you from playing next round.",
		},

		Respond: stopPlaying,
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionStartGame,
			Description: "Starts the game.",
		},

		Respond:     shared.StartGame,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionEndGame,
			Description: "Ends the game.",
		},
		Respond:     shared.EndGameInteraction,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        "day_time",
			Description: "Triggers day for the current game",
		},

		Respond:     triggerDay,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionNightTime,
			Description: "Triggers night for the current game",
		},

		Respond:     triggerNight,
		Authorizers: []lib.CommandAuthorizer{authorizors.IsAdmin},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionVote,
			Description: "Vote to hang. Leave off the target if you wish to unvote.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        lib.ActionOptionVoteUser,
					Description: "Select a player",
					Required:    false,
				},
			},
		},

		Respond: voteFor,
		Authorizers: []lib.CommandAuthorizer{
			authorizors.CharacterExists(lib.ActionOptionVoteUser),
			authorizors.IsAlive,
			canVote,
			authorizors.IsDayTime,
		},
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionWhoIsAlive,
			Description: "List who is alive and/or playing",
		},
		Respond: whoIsAlive,
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionShowVotesFor,
			Description: "List number of votes for each player",
		},
		Respond: showVotesFor,
	})

	cr.RegisterGlobal(lib.Command{
		ApplicationCommand: &discordgo.ApplicationCommand{
			Name:        lib.ActionShowVotersFor,
			Description: "List votes for each player",
		},
		Respond: showVotersFor,
	})

	return nil
}

func ping(ia *lib.InteractionArgs) error {
	return ia.Interaction.Respond("Pong!", false)
}

func aliveAndDeadList(alive, dead []*models.GuildCharacter, showAliveDescription, showDeadDescription bool) (memberList string) {
	memberList = "Alive:"
	for _, v := range alive {
		memberList += "\n  " + "<@" + v.Id + ">"
		if showAliveDescription {
			memberList += " - " + v.CharacterDescription()
		}
	}
	memberList += "\n\nDead:"
	for _, v := range dead {
		memberList += "\n  <@" + v.Id + ">"
		if showDeadDescription {
			memberList += " - " + v.CharacterDescription()
		}
	}
	if len(dead) == 0 {
		memberList += "\n  (none)"
	}
	return
}
