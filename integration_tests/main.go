package integration_tests

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/setup"
	"discord-werewolf/lib/shared"
	"discord-werewolf/lib/testlib"

	"github.com/samber/do"
)

func CallInteraction(args lib.SessionArgs, options testlib.TestInteractionOptions) {
	commandRegistrar := do.MustInvoke[*lib.CommandRegistrar](args.Injector)
	guild, err := args.AppGuild()
	if err != nil {
		panic(err)
	}
	commands := commandRegistrar.GetAllCommands(guild.Id)
	interaction := testlib.NewTestInteraction(args, options)
	interactionArgs := &lib.InteractionArgs{
		SessionArgs: args,
		Interaction: interaction,
	}
	shared.HandleInteraction(commands, interactionArgs)
}

func StartIntegratedTestGame(memberCount int, playingCount int, callback testlib.TestInitCallback) lib.SessionArgs {
	return testlib.StartTestGame(memberCount, playingCount, func(injector *do.Injector) {
		callback(injector)
		setup.SetupModules(injector)
	})
}

func StartDefaultIntegratedTestGame(memberCount int, playingCount int) lib.SessionArgs {
	return testlib.StartTestGame(memberCount, playingCount, func(injector *do.Injector) {
		setup.SetupModules(injector)
	})
}
