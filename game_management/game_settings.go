package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func changeGameMode(ia *lib.InteractionArgs) error {
	var err error

	data := ia.Interaction.MessageComponentData()
	selectedValue := data.Values[0]
	clearChannel := func() error {
		channel, err := ia.ChannelByAppId(models.ChannelAdminCharacters)
		if err != nil {
			return errors.Wrap(err, "could not get db channel")
		}
		err = ia.Session.ClearChannelMessagesUnless(channel.Id, gameSettingsMessageKeeper)
		if err != nil {
			return errors.Wrap(err, "Could not clear characters channel")
		}
		return nil
	}
	switch selectedValue {
	case models.MessageGameMode_BalancedTeams:
		if err = clearChannel(); err != nil {
			return err
		}
	case models.MessageGameMode_BalancedSelect:
		if err = clearChannel(); err != nil {
			return err
		}
	case models.MessageGameMode_RandomTeams:
		if err = clearChannel(); err != nil {
			return err
		}
	case models.MessageGameMode_RandomSelect:
		if err = clearChannel(); err != nil {
			return err
		}
	default:
		_ = ia.Interaction.Respond("A server error occurred", true)
		return errors.New("unknown game mode: " + selectedValue)
	}
	//guild, err := ia.AppGuild()
	//if err != nil {
	//	return errors.Wrap(err, "could not get game guild")
	//}

	log.Println(selectedValue)
	return nil
}

func gameSettingsMessageKeeper(message *discordgo.Message) (bool, error) {
	return messageComponentExists(models.MessageGameMode, message.Components), nil
}

func messageComponentExists(id string, components []discordgo.MessageComponent) bool {
	if components == nil {
		return false
	}
	for _, component := range components {
		if menu, ok := component.(discordgo.SelectMenu); ok {
			if menu.CustomID == id {
				return true
			}
		}
		if row, ok := component.(discordgo.ActionsRow); ok {
			if messageComponentExists(id, row.Components) {
				return true
			}
		}
	}
	return false
}
