package shared

import (
	"discord-werewolf/lib"
	"fmt"
	"log"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/pkg/errors"
)

func responseRecover(args *lib.InteractionArgs) {
	p := recover()
	if p == nil {
		return
	}
	if err, ok := p.(error); ok {
		errorRespond(args.Interaction, err)
	} else if str, ok := p.(string); ok {
		errorRespond(args.Interaction, str)
		fmt.Println(debug.Stack())
	} else {
		log.Println("Unknown panic recovered: " + reflect.TypeOf(p).Name())
		fmt.Println(debug.Stack())
	}
}

func authorize(args *lib.InteractionArgs, authorizers []lib.Authorizer) error {
	for _, auth := range authorizers {
		err := auth(args)
		if err != nil {
			if _, ok := err.(lib.PermissionDeniedError); ok {
				if err = args.Interaction.Respond(err.Error(), true); err != nil {
					log.Println(err)
				}
			} else {
				errorRespond(args.Interaction, fmt.Sprintf("Could not authorize command: %s", err.Error()))
			}
			return err
		}
	}
	return nil
}

func HandleAction(actions map[string]lib.SettingAction, args *lib.InteractionArgs) {
	defer responseRecover(args)
	var err error
	name := args.Interaction.MessageComponentData().CustomID
	if action, ok := actions[name]; ok {
		if err = authorize(args, action.Authorizers); err != nil {
			return
		}
		if action.Respond == nil {
			errorRespond(args.Interaction, fmt.Sprintf("Action has no Respond method: %s", name))
			return
		}
		if err = action.Respond(args); err != nil {
			errorRespond(args.Interaction, err)
			return
		}
	} else {
		errorRespond(args.Interaction, fmt.Sprintf("Unknown action: %s", name))
	}
}

func HandleCommand(commands map[string]lib.Command, args *lib.InteractionArgs) {
	defer responseRecover(args)
	var err error
	commandName := args.Interaction.CommandData().Name
	if cmd, ok := commands[commandName]; ok {
		if err = authorize(args, cmd.Authorizers); err != nil {
			return
		}
		if cmd.Respond == nil {
			errorRespond(args.Interaction, fmt.Sprintf("Command has no Respond method: %s", commandName))
			return
		}
		if err = cmd.Respond(args); err != nil {
			errorRespond(args.Interaction, err)
			return
		}
	} else {
		errorRespond(args.Interaction, fmt.Sprintf("Unknown command: %s", commandName))
	}
}

func printStackTrace(err lib.StackTracer) {
	var buf strings.Builder
	buf.WriteString(err.Error() + "\n")
	for _, frame := range err.StackTrace() {
		buf.WriteString(fmt.Sprintf("%+v\n", frame))
	}
	log.Println(buf.String())
}

func errorRespond(i lib.Interaction, message interface{}) {
	if err, ok := message.(error); ok {
		var tracer lib.StackTracer
		for err != nil {
			if stackErr, ok := err.(lib.StackTracer); ok {
				tracer = stackErr
			}
			err = errors.Unwrap(err)
		}
		if tracer != nil {
			printStackTrace(tracer)
		} else {
			log.Println(err)
		}
	} else {
		log.Println(message)
	}
	_ = i.Respond("There was a system error.", true)
}
