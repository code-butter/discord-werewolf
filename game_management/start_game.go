package game_management

import (
	"discord-werewolf/lib"
)

// TODO: implement different game modes
// TODO: enable scheduled start
// This is public for tests in other packages

func playing(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.AssignRoleToRequester(lib.RolePlaying); err != nil {
		return err
	}
	return ia.Interaction.Respond("Now playing!", false)
}

func stopPlaying(ia *lib.InteractionArgs) error {
	if err := ia.Interaction.RemoveRoleFromRequester(lib.RolePlaying); err != nil {
		return err
	}
	return ia.Interaction.Respond("Stopped playing.", false)
}
