package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"embed"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"discord-werewolf/lib"

	"github.com/bwmarrin/discordgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *sql.DB
var discordClient *discordgo.Session
var ctx = context.Background() // TODO: make this listen to signals

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {

	// Env vars
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable not set")
	}

	clientId := os.Getenv("CLIENT_ID")
	if clientId == "" {
		log.Fatal("CLIENT_ID environment variable not set")
	}

	var err error

	// Database
	if db, err = sql.Open("sqlite3", os.Args[1]+"?_rt=on&_fk=on"); err != nil {
		log.Fatal(errors.Wrap(err, "Could not connect to database"))
	}
	defer db.Close()

	// TODO: move this to an "upgrade" subcommand
	if err = goose.SetDialect("sqlite3"); err != nil {
		log.Fatal(errors.Wrap(err, "Could not set goose dialect"))
	}

	// TODO: check if migrations are needed and back up database
	goose.SetBaseFS(embedMigrations)
	if err = goose.Up(db, "migrations"); err != nil {
		log.Fatal(errors.Wrap(err, "Could not do auto-migrations"))
	}

	gormDB, err := gorm.Open(sqlite.New(sqlite.Config{
		Conn: db,
	}))
	if err != nil {
		log.Fatal(errors.Wrap(err, "Could not connect to database with Gorm"))
	}

	// Set up global services
	lib.DB = db
	lib.Ctx = ctx
	lib.GormDB = gormDB

	// Setup Discord
	if discordClient, err = discordgo.New("Bot " + token); err != nil {
		log.Fatal(err)
	}

	if err = game_management.Init(); err != nil {
		log.Fatal(err)
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
		log.Fatal(errors.Wrap(err, "Could not connect to Discord"))
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
			log.Fatal(errors.Wrap(err, "Could not bulk overwrite global commands"))
		}
	}

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
