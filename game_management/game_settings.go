package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/characters"
	"discord-werewolf/lib/models"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const unchecked = "⬜"
const checked = "☑️"

func createTeams(channelId string, ia *lib.InteractionArgs) (err error) {
	for _, team := range characters.Teams {
		err = ia.Session.MessageComplex(channelId, &discordgo.MessageSend{
			Components: []discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.Button{
							Label:    team.Label,
							Emoji:    &discordgo.ComponentEmoji{Name: unchecked},
							CustomID: models.MessageTeam + ":" + team.Id,
							Style:    discordgo.PrimaryButton,
						},
					},
				},
			},
		})
		if err != nil {
			return errors.Wrap(err, "could not send message on admin character channel")
		}
	}
	return
}

func createCharacters(channelId string, ia *lib.InteractionArgs, includeCount bool) (err error) {
	for _, team := range characters.Teams {
		err = ia.Session.Message(channelId, "## "+team.Label)
		if err != nil {
			return errors.Wrap(err, "could not send message on admin character channel")
		}
		var rows []discordgo.MessageComponent

		sendRows := func() (err error) {
			err = ia.Session.MessageComplex(channelId, &discordgo.MessageSend{
				Components: rows,
			})
			if err != nil {
				return
			}
			time.Sleep(1100 * time.Millisecond) // avoid Discord rate limiting
			rows = nil
			return
		}
		for i, character := range team.Characters {
			msgComponents := []discordgo.MessageComponent{
				&discordgo.Button{
					Label:    character.Label,
					Emoji:    &discordgo.ComponentEmoji{Name: unchecked},
					CustomID: models.MessageCharacterToggle + ":" + character.Id,
					Style:    discordgo.PrimaryButton,
				},
			}
			if includeCount {
				msgComponents = append(msgComponents, &discordgo.Button{
					Label:    "Count: 0",
					CustomID: models.MessageCharacterCount + ":" + character.Id,
					Style:    discordgo.SecondaryButton,
				})
			}
			rows = append(rows, &discordgo.ActionsRow{Components: msgComponents})
			if (i+1)%5 == 0 {
				if err = sendRows(); err != nil {
					return errors.Wrap(err, "could not send message on admin character channel")
				}
			}
		}
		if len(rows) > 0 {
			if err = sendRows(); err != nil {
				return errors.Wrap(err, "could not send message on admin character channel")
			}
		}
	}
	return
}

func changeGameMode(ia *lib.InteractionArgs) (err error) {
	if err = ia.Interaction.SilentDeferred(); err != nil {
		return err
	}
	data := ia.Interaction.MessageComponentData()
	selectedValue := data.Values[0]

	channel, err := ia.ChannelByAppId(models.ChannelAdminCharacters)
	if err != nil {
		return errors.Wrap(err, "could not get admin characters db channel")
	}
	err = ia.Session.ClearChannelMessagesUnless(channel.Id, gameSettingsMessageKeeper)
	if err != nil {
		return errors.Wrap(err, "could not clear characters channel")
	}

	switch selectedValue {
	case models.MessageGameMode_BalancedTeams:
		if err = createTeams(channel.Id, ia); err != nil {
			return
		}
	case models.MessageGameMode_BalancedSelect:
		if err = createCharacters(channel.Id, ia, true); err != nil {
			return
		}
	case models.MessageGameMode_RandomTeams:
		if err = createTeams(channel.Id, ia); err != nil {
			return
		}
	case models.MessageGameMode_RandomSelect:
		if err = createCharacters(channel.Id, ia, false); err != nil {
			return
		}
	default:
		err = errors.New("unknown game mode: " + selectedValue)
		return
	}

	err = ia.Interaction.FollowupMessageEdit(discordgo.MessageEdit{
		Components: selectOptions(data.Values, ia.Interaction.Message().Components),
	})
	return
}

func changeCharacterCount(args *lib.InteractionArgs) error {
	return args.Interaction.InteractionRespond(&discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseModal,
		Data: &discordgo.InteractionResponseData{
			CustomID: models.ModalCharacterCount + ":" + args.Interaction.Message().ID + ":" + args.Interaction.MessageComponentData().CustomID,
			Title:    "Change character count",
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.TextInput{
							CustomID: "character-count",
							Label:    "Value",
							Style:    discordgo.TextInputShort,
						},
					},
				},
			},
		},
	})
}

func handleCharacterCount(args *lib.InteractionArgs) (err error) {
	parts := strings.SplitN(args.Interaction.CustomIdData(), ":", 2)
	messageId := parts[0]
	buttonCustomId := parts[1]
	characterChannel, err := args.ChannelByAppId(models.ChannelAdminCharacters)
	if err != nil {
		return
	}
	value, err := inputValue("character-count", args.Interaction.ModelSubmitData().Components)
	if err != nil {
		return errors.Wrap(err, "could not get value from modal")
	}
	message, err := args.Session.GetMessage(characterChannel.Id, messageId)
	if err != nil {
		return errors.Wrap(err, "could not get message from passed settings: "+messageId+":"+buttonCustomId)
	}
	button, err := getButton(buttonCustomId, message.Components)
	if err != nil {
		return errors.Wrap(err, "could not get calling button from message: "+messageId+":"+buttonCustomId)
	}
	button.Label = "Count: " + value
	err = args.Session.MessageEditComplex(&discordgo.MessageEdit{
		Components: &message.Components,
		ID:         message.ID,
		Channel:    message.ChannelID,
	})
	if err != nil {
		return errors.Wrap(err, "could not edit message")
	}
	if err = args.Interaction.SilentDeferred(); err != nil {
		return errors.Wrap(err, "could not respond to interaction")
	}
	return
}

func messageChangeButton(ia *lib.InteractionArgs) (err error) {
	customId := ia.Interaction.MessageComponentData().CustomID
	err = ia.Interaction.FollowupMessageEdit(discordgo.MessageEdit{
		Components: toggleButton(customId, ia.Interaction.Message().Components),
	})
	if err = ia.Interaction.SilentDeferred(); err != nil {
		return
	}
	return
}

func gameSettingsMessageKeeper(message *discordgo.Message) (bool, error) {
	return messageComponentExists(models.MessageGameMode, message.Components), nil
}

func inputValue(id string, components []discordgo.MessageComponent) (string, error) {
	for _, component := range components {
		if row, ok := component.(*discordgo.ActionsRow); ok {
			if value, err := inputValue(id, row.Components); err != nil || value != "" {
				return value, err
			}
		}
		if input, ok := component.(*discordgo.TextInput); ok {
			if input.CustomID == id {
				return input.Value, nil
			}
		}
	}
	return "", fmt.Errorf("input with id %s not found", id)
}

func getButton(id string, components []discordgo.MessageComponent) (*discordgo.Button, error) {
	for _, component := range components {
		if row, ok := component.(*discordgo.ActionsRow); ok {
			if button, err := getButton(id, row.Components); err == nil {
				return button, nil
			}
		}
		if button, ok := component.(*discordgo.Button); ok {
			if button.CustomID == id {
				return button, nil
			}
		}
	}
	return nil, fmt.Errorf("button with id %s not found", id)
}

func toggleButton(id string, components []discordgo.MessageComponent) *[]discordgo.MessageComponent {
	for _, component := range components {
		if row, ok := component.(*discordgo.ActionsRow); ok {
			toggleButton(id, row.Components)
		}
		if button, ok := component.(*discordgo.Button); ok {
			if button.CustomID != id {
				continue
			}
			if button.Emoji.Name == checked {
				button.Emoji.Name = unchecked
			} else {
				button.Emoji.Name = checked
			}
		}
	}
	return &components
}

func selectOptions(values []string, components []discordgo.MessageComponent) *[]discordgo.MessageComponent {
	for _, component := range components {
		if row, ok := component.(*discordgo.ActionsRow); ok {
			selectOptions(values, row.Components)
		}
		if menu, ok := component.(*discordgo.SelectMenu); ok {
			for i, option := range menu.Options {
				found := false
				for _, value := range values {
					if value == option.Value {
						found = true
					}
				}
				option.Default = found
				menu.Options[i] = option
			}
		}
	}
	return &components
}

func messageComponentExists(id string, components []discordgo.MessageComponent) bool {
	if components == nil {
		return false
	}
	for _, component := range components {
		if menu, ok := component.(*discordgo.SelectMenu); ok {
			if menu.CustomID == id {
				return true
			}
		}
		if row, ok := component.(*discordgo.ActionsRow); ok {
			if messageComponentExists(id, row.Components) {
				return true
			}
		}
	}
	return false
}
