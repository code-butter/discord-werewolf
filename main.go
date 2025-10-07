package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"discord-werewolf/guild_management"
	"discord-werewolf/lib"
	"discord-werewolf/werewolves"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	ctx := context.Background() // TODO: tie this to signals

	// Env vars
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		lib.Fatal("DISCORD_TOKEN environment variable not set")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		lib.Fatal("CLIENT_ID environment variable not set")
	}

	var err error
	// Database
	var db *sql.DB
	if db, err = sql.Open("sqlite3", os.Args[1]+"?_rt=on&_fk=on"); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to database"))
	}
	defer db.Close()
	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}))
	if err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to database with Gorm"))
	}

	// TODO: move this to an "upgrade" subcommand
	if err = lib.MigrateUp(db); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not apply migrations"))
	}

	// Setup Discord
	var discordClient *discordgo.Session
	if discordClient, err = discordgo.New("Bot " + token); err != nil {
		lib.Fatal(err)
	}

	if err = game_management.Setup(); err != nil {
		lib.Fatal(err)
	}
	if err = guild_management.Setup(); err != nil {
		lib.Fatal(err)
	}
	if err = werewolves.Setup(); err != nil {
		lib.Fatal(err)
	}

	commands := lib.GetCommands()

	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var err error
		if i.Type == discordgo.InteractionApplicationCommand {
			commandName := i.ApplicationCommandData().Name
			interaction := lib.NewLiveInteraction(s, i)
			if cmd, ok := commands[commandName]; ok {
				interactionArgs := &lib.InteractionArgs{
					ServiceArgs: &lib.ServiceArgs{
						Session: lib.NewLiveDiscordSession(i.GuildID, s),
						GormDB:  gormDB,
						Ctx:     ctx,
					},
					Interaction: interaction,
				}
				if len(cmd.Authorizers) > 0 {
					authorized := false
					for _, auth := range cmd.Authorizers {
						a, err := auth(interactionArgs)
						if err != nil {
							errorRespond(interaction, fmt.Sprintf("Could not authorize command: %s", err.Error()))
							return
						}
						if a {
							authorized = true
							break
						}
					}
					if !authorized {
						if err = interactionArgs.Interaction.Respond("unauthorized", true); err != nil {
							log.Println(err)
						}
					}
				}
				if cmd.Respond == nil {
					errorRespond(interaction, fmt.Sprintf("Command has no Respond method: %s\n", commandName))
					return
				}
				if err = cmd.Respond(interactionArgs); err != nil {
					errorRespond(interaction, err.Error())
					return
				}
			} else {
				errorRespond(interaction, fmt.Sprintf("Unknown command: %s", commandName))
				return
			}
		}
	})
	if err = discordClient.Open(); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to Discord"))
	}
	defer discordClient.Close()

	// TODO: move this to an "upgrade" subcommand
	globalCommands := make([]*discordgo.ApplicationCommand, 0)
	for _, cmd := range commands {
		if cmd.Global {
			globalCommands = append(globalCommands, cmd.ApplicationCommand)
		}
	}
	if len(globalCommands) > 0 {
		_, err = discordClient.ApplicationCommandBulkOverwrite(clientId, "", globalCommands)
		if err != nil {
			lib.Fatal(errors.Wrap(err, "Could not bulk overwrite global commands"))
		}
	}

	currentTz, err := lib.SystemTimeZone()
	if err != nil {
		lib.Fatal(errors.Wrap(err, "Could not get system time zone"))
	}

	log.Printf("System timezone: %s\n", currentTz)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func errorRespond(i lib.Interaction, message string) {
	log.Println(message)
	_ = i.Respond("There was a system error.", true)
}
