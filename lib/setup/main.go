package setup

import (
	"discord-werewolf/game_management"
	"discord-werewolf/lib/shared"
	"discord-werewolf/werewolves"
	"log"

	"github.com/samber/do"
)

func SetupModules(injector *do.Injector) {
	// Setup modules. Keep in mind that order is important here. Callbacks will be run in the order they
	// are set up in these functions.
	setupFunctions := []shared.SetupFunction{
		game_management.Setup,
		werewolves.Setup,
	}
	for _, f := range setupFunctions {
		if err := f(injector); err != nil {
			log.Fatal(err)
		}
	}
}
