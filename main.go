package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-werewolf/lib"

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
		fatal("DISCORD_TOKEN environment variable not set")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		fatal("CLIENT_ID environment variable not set")
	}

	var err error

	// Database
	var db *sql.DB
	if db, err = sql.Open("sqlite3", os.Args[1]+"?_rt=on&_fk=on"); err != nil {
		fatal(errors.Wrap(err, "Could not connect to database"))
	}
	defer db.Close()
	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}))
	if err != nil {
		fatal(errors.Wrap(err, "Could not connect to database with Gorm"))
	}

	// Set up global services
	lib.DB = db
	lib.Ctx = ctx
	lib.GormDB = gormDB

	// TODO: move this to an "upgrade" subcommand
	lib.MigrateUp()

	// Setup Discord
	// TODO: maybe make a wrapper service for mock testing?
	var discordClient *discordgo.Session
	if discordClient, err = discordgo.New("Bot " + token); err != nil {
		fatal(err)
	}

	if err = game_management.Setup(); err != nil {
		fatal(err)
	}

	commands := lib.GetCommands()

	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		var err error
		if i.Type == discordgo.InteractionApplicationCommand {
			interaction := lib.LiveInteraction{
				Session:           s,
				InteractionCreate: i,
			}
			commandName := i.ApplicationCommandData().Name
			if cmd, ok := commands[commandName]; ok {
				if cmd.Respond == nil {
					log.Printf("Command has no Respond method: %s\n", commandName)
				}
				if err = cmd.Respond(interaction); err != nil {
					log.Println(err)
				}
			} else {
				msg := fmt.Sprintf("Unknown command: %s", commandName)
				log.Println(msg)
				err = interaction.Respond(msg, true)
				if err != nil {
					log.Println(err)
				}
			}
		}
	})
	if err = discordClient.Open(); err != nil {
		fatal(errors.Wrap(err, "Could not connect to Discord"))
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
			fatal(errors.Wrap(err, "Could not bulk overwrite global commands"))
		}
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func fatal(err interface{}) {
	log.Printf("%+v\n", err)
	os.Exit(1)
}
