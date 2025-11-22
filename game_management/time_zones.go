package game_management

import (
	"discord-werewolf/lib"
	"fmt"
	"regexp"
	"strings"

	"github.com/samber/do"
)

func getTimeZones(ia *lib.InteractionArgs) error {
	var err error
	if err = ia.Interaction.DeferredResponse("Loading timezones...", true); err != nil {
		return err
	}
	areaName := ia.Interaction.CommandData().GetOption("area").Value.(string)
	matcher := regexp.MustCompile(fmt.Sprintf("^%s/", areaName))
	tzs := lib.AllTimeZoneNames()
	builder := strings.Builder{}
	for _, tz := range tzs {
		if matcher.MatchString(tz) {
			builder.WriteString(fmt.Sprintln(tz))
			if builder.Len() > 1900 { // discord max message length is 2000 characters
				if err = ia.Interaction.FollowupMessage(builder.String(), true); err != nil {
					return err
				}
				builder.Reset()
			}
		}
	}
	return ia.Interaction.FollowupMessage(builder.String(), true)
}

func setTimeZone(ia *lib.InteractionArgs) error {
	settings := do.MustInvoke[*lib.GuildSettings](ia.Injector)
	tzName := ia.Interaction.CommandData().GetOption("timezone").Value.(string)
	if err := settings.SetTimeZone(ia.Interaction.GuildId(), tzName); err != nil {
		_ = ia.Interaction.Respond("Unable to set timezone", true)
		return err
	}
	return ia.Interaction.Respond(fmt.Sprintf("Set timezone to %s", tzName), true)

}
