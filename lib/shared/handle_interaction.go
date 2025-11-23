package shared

import (
	"discord-werewolf/lib"
	"fmt"
	"log"
	"runtime/debug"
)

func HandleInteraction(commands map[string]lib.Command, args *lib.InteractionArgs) {
	defer func() {
		p := recover()
		if p != nil {
			if err, ok := p.(error); ok {
				errorRespond(args.Interaction, err.Error())
			} else if str, ok := p.(string); ok {
				errorRespond(args.Interaction, str)
			}
			fmt.Println(debug.Stack())
		}
	}()
	var err error
	commandName := args.Interaction.CommandData().Name
	if cmd, ok := commands[commandName]; ok {
		for _, auth := range cmd.Authorizers {
			err := auth(args)
			if err != nil {
				if _, ok := err.(lib.PermissionDeniedError); ok {
					if err = args.Interaction.Respond(err.Error(), true); err != nil {
						log.Println(err)
					}
				} else {
					errorRespond(args.Interaction, fmt.Sprintf("Could not authorize command: %s", err.Error()))
				}
				return
			}
		}
		if cmd.Respond == nil {
			errorRespond(args.Interaction, fmt.Sprintf("Command has no Respond method: %s\n", commandName))
			return
		}
		if err = cmd.Respond(args); err != nil {
			errorRespond(args.Interaction, err.Error())
			return
		}
	} else {
		errorRespond(args.Interaction, fmt.Sprintf("Unknown command: %s", commandName))
		return
	}
}

func errorRespond(i lib.Interaction, message string) {
	log.Println(message)
	_ = i.Respond("There was a system error.", true)
}
