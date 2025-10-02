package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"discord-werewolf/guild_management"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if err = guild_management.Setup(); err != nil {
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
				if len(cmd.Authorizers) > 0 {
					authorized := false
					for _, auth := range cmd.Authorizers {
						a, err := auth(interaction)
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
						if err = interaction.Respond("unauthorized", true); err != nil {
							log.Println(err)
						}
					}
				}
				if cmd.Respond == nil {
					errorRespond(interaction, fmt.Sprintf("Command has no Respond method: %s\n", commandName))
					return
				}
				if err = cmd.Respond(interaction); err != nil {
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

	currentTz, err := lib.SystemTimeZone()
	if err != nil {
		fatal(errors.Wrap(err, "Could not get system time zone"))
	}

	log.Printf("System timezone: %s\n", currentTz)

	log.Println("Bot is now running.  Press CTRL-C to exit.")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

func TimedDayNight(ctx context.Context, session *discordgo.Session) {
	var err error

	systemTz, err := lib.SystemTimeZone()

	if err != nil {
		log.Println(err)
		systemTz = time.UTC
	}

	for ctx.Err() == nil {
		cutoff := time.Now().UTC().Add(-23*time.Hour + 59*time.Minute)
		now := time.Now().UTC()
		var guilds []models.Guild
		if result := lib.GormDB.Where("game_going = 1 AND paused = 0").Find(&guilds); result.Error != nil {
			panic(result.Error)
		}
		var finishedGuildIds []string
		for _, guild := range guilds {
			lastCycleRan := guild.LastCycleRan
			if lastCycleRan == nil {
				lastCycleRan = &now
			}
			if lastCycleRan.Before(cutoff) {
				if guild.DayNight {
					err = game_management.StartNight(guild.Id, session)
				} else {
					err = game_management.StartDay(guild.Id, session)
				}
				if err != nil {
					log.Println(err)
					continue
				}
				finishedGuildIds = append(finishedGuildIds, guild.Id)
				continue
			}
			var guildTz *time.Location
			if guild.TimeZone == "" {
				guildTz = systemTz
			} else {
				guildTz, err = time.LoadLocation(guild.TimeZone)
				if err != nil {
					log.Println(err)
					guildTz = systemTz
				}
			}
			if guild.DayNight {
				nightTime := guild.NightTime
				if nightTime == nil {
					newNightTime := time.Date(0, 0, 0, 18, 0, 0, 0, guildTz).UTC()
					nightTime = &models.TimeOnly{Time: &newNightTime}
				}
				if nightTime.BeforeOrOn(now) && nightTime.AfterOrOn(*lastCycleRan) {
					if err = game_management.StartNight(guild.Id, session); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			} else {
				dayTime := guild.DayTime
				if dayTime == nil {
					newDayTime := time.Date(0, 0, 0, 6, 0, 0, 0, guildTz).UTC()
					dayTime = &models.TimeOnly{Time: &newDayTime}
				}
				if dayTime.BeforeOrOn(now) && dayTime.AfterOrOn(*lastCycleRan) {
					if err = game_management.StartDay(guild.Id, session); err != nil {
						log.Println(err)
						continue
					}
					finishedGuildIds = append(finishedGuildIds, guild.Id)
				}
			}
		}
		_, err = gorm.G[models.Guild](lib.GormDB).
			Where("id in ?", finishedGuildIds).
			Update(lib.Ctx, "last_cycle_ran", now)
		if err != nil {
			fatal(err)
		}
		time.Sleep(time.Minute * 5)
	}
}

func errorRespond(i lib.LiveInteraction, message string) {
	log.Println(message)
	_ = i.Respond("There was a system error.", true)
}

func fatal(err interface{}) {
	log.Printf("%+v\n", err)
	os.Exit(1)
}
