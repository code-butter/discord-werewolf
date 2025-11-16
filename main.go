package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"discord-werewolf/lib"
	"discord-werewolf/lib/shared"
	"discord-werewolf/werewolves"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		lib.Fatal("DISCORD_TOKEN environment variable not set")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		lib.Fatal("CLIENT_ID environment variable not set")
	}

	injector := shared.Setup()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	do.ProvideValue[context.Context](injector, ctx)

	var err error
	// Database
	var db *sql.DB
	if db, err = sql.Open("sqlite3", os.Args[1]+"?_rt=on&_fk=on"); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to database"))
	}
	defer db.Close()

	do.ProvideValue[*sql.DB](injector, db)

	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}))
	if err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to database with Gorm"))
	}

	do.ProvideValue[*gorm.DB](injector, gormDB)

	// TODO: move this to an "upgrade" subcommand
	if err = lib.MigrateUp(db); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not apply migrations"))
	}

	// Setup Discord
	var discordClient *discordgo.Session
	if discordClient, err = discordgo.New("Bot " + token); err != nil {
		lib.Fatal(err)
	}
	lib.SetBotConnection(discordClient)

	do.ProvideValue[*discordgo.Session](injector, discordClient)

	clock := lib.RealClock{}

	do.ProvideValue[lib.Clock](injector, clock)

	// Setup sections. Keep in mind that order is important here. Callbacks will be run in the order they
	// are set up in these functions.
	if err = game_management.Setup(injector); err != nil {
		lib.Fatal(err)
	}
	if err = werewolves.Setup(injector); err != nil {
		lib.Fatal(err)
	}

	// Discord handlers
	commands := lib.GetGlobalCommands()
	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var err error
		if i.Type == discordgo.InteractionApplicationCommand {
			commandName := i.ApplicationCommandData().Name
			interaction := lib.NewLiveInteraction(i)
			if cmd, ok := commands[commandName]; ok {
				interactionArgs := &lib.InteractionArgs{
					SessionArgs: lib.SessionArgs{
						Session:  lib.GetGuildDiscordSession(i.GuildID),
						Injector: injector,
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
		globalCommands = append(globalCommands, cmd.ApplicationCommand)
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

	// Start timers
	go game_management.TimedDayNight(injector, time.Second*15)

	log.Printf("System timezone: %s\n", currentTz)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	<-ctx.Done()
	log.Println("Shutting down...")
	time.Sleep(5 * time.Second)
}

func errorRespond(i lib.Interaction, message string) {
	log.Println(message)
	_ = i.Respond("There was a system error.", true)
}
