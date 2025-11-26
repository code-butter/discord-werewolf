/*
Copyright (C) 2025  Jeremy Nicoll

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/

package main

import (
	"context"
	"database/sql"
	"discord-werewolf/game_management"
	"discord-werewolf/lib"
	"discord-werewolf/lib/setup"
	"discord-werewolf/lib/shared"
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

	injector := shared.SetupInjector()

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

	clock := lib.RealClock{}

	do.ProvideValue[lib.Clock](injector, clock)

	// Setup Discord
	var discordClient *discordgo.Session
	if discordClient, err = discordgo.New("Bot " + token); err != nil {
		lib.Fatal(err)
	}

	var sessionGetter lib.GuildSessionGetter
	sessionGetter = func(guildId string) (lib.DiscordSession, error) {
		return lib.NewGuildDiscordSession(guildId, discordClient, 3*time.Minute, clock), nil
	}
	sessionProvider := lib.NewDiscordSessionProvider(sessionGetter)
	do.ProvideValue[lib.DiscordSessionProvider](injector, sessionProvider)

	setup.SetupModules(injector)

	commandRegistrar := do.MustInvoke[*lib.CommandRegistrar](injector)

	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			commands := commandRegistrar.GetAllCommands(i.GuildID)
			session := sessionProvider.GetSession(i.GuildID)
			interaction := lib.NewLiveInteraction(i, session)
			args := &lib.InteractionArgs{
				SessionArgs: &lib.SessionArgs{
					Session:  session,
					Injector: injector,
				},
				Interaction: interaction,
			}
			shared.HandleInteraction(commands, args)
		}
	})
	if err = discordClient.Open(); err != nil {
		lib.Fatal(errors.Wrap(err, "Could not connect to Discord"))
	}
	defer discordClient.Close()

	// TODO: move this to an "upgrade" subcommand
	globalCommands := make([]*discordgo.ApplicationCommand, 0)
	for _, cmd := range commandRegistrar.GetGlobalCommands() {
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
