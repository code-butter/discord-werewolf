package shared

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var InitialChannels map[string]models.GuildChannel

func init() {
	InitialChannels = map[string]models.GuildChannel{
		"game-instructions": {
			Name:     "Game Instructions",
			AppId:    models.CatChannelInstructions,
			Children: &[]models.GuildChannel{},
		},
		// seer and lovers channels are created on demand
		"the-town": {
			Name:  "The Town",
			AppId: models.CatChannelTheTown,
			Children: &[]models.GuildChannel{
				{
					Name:  models.ChannelTownSquare,
					AppId: models.ChannelTownSquare,
				},
				{
					Name:  models.ChannelWerewolves,
					AppId: models.ChannelWerewolves,
				},
				{
					Name:  models.ChannelWitch,
					AppId: models.ChannelWitch,
				},
				{
					Name:  models.ChannelMasons,
					AppId: models.ChannelMasons,
				},
				{
					Name:  models.ChannelBodyguard,
					AppId: models.ChannelBodyguard,
				},
				{
					Name:  models.ChannelAfterLife,
					AppId: models.ChannelAfterLife,
				},
			},
		},
		"admin": {
			Name:     "Admin",
			AppId:    models.CatChannelAdmin,
			Children: &[]models.GuildChannel{},
		},
	}
}

// Setup is shared between tests and the application
func Setup() *do.Injector {
	injector := do.New()
	do.ProvideValue[*lib.GameListeners](injector, lib.NewGameListeners())
	do.Provide[*lib.GuildSettings](injector, lib.NewGameSettings)
	return injector
}

func InitGuild(ia *lib.InteractionArgs) error {
	var err error

	gormDB := do.MustInvoke[*gorm.DB](ia.Injector)

	if err = ia.Interaction.DeferredResponse("Initializing server...", true); err != nil {
		return errors.Wrap(err, "Could not send deferred response to Discord")
	}

	defer func() {
		if err != nil {
			_ = ia.Interaction.FollowupMessage("There was an error setting up the server.", true)
		} else {
			_ = ia.Interaction.FollowupMessage("Server initialized!", true)
		}
	}()

	guild, err := ia.Session.Guild()
	if err != nil {
		return errors.Wrap(err, "Could not get current guild")
	}

	var guildRecord *models.Guild
	if result := gormDB.Where("id = ?", guild.ID).First(&guildRecord); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			guildRecord = nil
		} else {
			return errors.Wrap(result.Error, "Could not get guild record")
		}
	}
	if guildRecord == nil {
		guildRecord = &models.Guild{
			Name: guild.Name,
			Id:   guild.ID,
		}
	}

	////////////////////////////////////////
	// TODO: REMOVE/MODIFY THIS BEFORE PROD DEPLOY
	discordChannels, err := ia.Session.Channels()

	for _, channel := range discordChannels {
		if channel.Name == "general" {
			continue
		}
		if err = ia.Session.DeleteChannel(channel.ID); err != nil {
			return errors.Wrap(err, "Could not delete channel")
		}
	}
	////////////////////////////////////////

	// Create roles if they don't exist
	guildRoles, err := ia.Session.GetRoles()
	if err != nil {
		return errors.Wrap(err, "Could not get guild roles")
	}
	for _, role := range lib.Roles {
		if err = ia.Session.EnsureRoleCreated(role.Name, role.Color, guildRoles); err != nil {
			return errors.Wrap(err, fmt.Sprintf("Could not create role %s", role.Name))
		}
	}

	// Map roles
	roles, err := ia.Session.GetRoles()
	if err != nil {
		return errors.Wrap(err, "Could not get guild roles")
	}
	roleMap := make(map[string]*discordgo.Role)
	for _, role := range roles {
		roleMap[role.Name] = role
	}

	saveChannels := models.GuildChannels{}
	for _, initChannel := range InitialChannels {
		discordCat, err := ia.Session.CreateCategoryChannel(initChannel.Name)
		if err != nil {
			return errors.Wrap(err, "Could not create category channel")
		}
		if initChannel.AppId == "admin" {
			if err = ia.Session.RoleChannelPermissions(discordCat.ID, roleMap["Admin"].ID, discordgo.PermissionViewChannel, 0); err != nil {
				return errors.Wrap(err, "Could not set channel permissions for admin role")
			}
			if err = ia.Session.RoleChannelPermissions(discordCat.ID, guild.ID, 0, discordgo.PermissionViewChannel); err != nil {
				return errors.Wrap(err, "Could not set channel permissions for admin role")
			}
		}
		cat := models.GuildChannel{
			Name:     initChannel.Name,
			Id:       discordCat.ID,
			AppId:    initChannel.AppId,
			Children: &[]models.GuildChannel{},
		}
		var catChildren []models.GuildChannel
		for _, child := range *initChannel.Children {
			discordChannel, err := ia.Session.CreateTextChannel(child.Name, cat.Id)
			if err != nil {
				msg := fmt.Sprintf("Could not create text channel for child %s in guild %s", child.Name, guild.Name)
				return errors.Wrap(err, msg)
			}
			if err = ia.Session.RoleChannelPermissions(discordChannel.ID, guild.ID, 0, discordgo.PermissionSendMessages); err != nil {
				return errors.Wrap(err, "Could not set universal channel permissions for channel "+child.Name)
			}
			if child.AppId == models.ChannelTownSquare {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Alive"].ID, discordgo.PermissionSendMessages|discordgo.PermissionViewChannel, 0); err != nil {
					return errors.Wrap(err, "Could not set Alive channel permissions for channel "+child.Name)
				}
			} else {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Alive"].ID, 0, discordgo.PermissionViewChannel); err != nil {
					return errors.Wrap(err, "Could not set Alive channel permissions for channel "+child.Name)
				}
			}
			if child.AppId == models.ChannelAfterLife {
				if err = ia.Session.RoleChannelPermissions(discordChannel.ID, roleMap["Dead"].ID, discordgo.PermissionSendMessages|discordgo.PermissionViewChannel, 0); err != nil {
					return errors.Wrap(err, "Could not set Dead channel permissions for channel "+child.Name)
				}
			}
			catChildren = append(catChildren, models.GuildChannel{
				Name:  child.Name,
				AppId: child.AppId,
				Id:    discordChannel.ID,
			})
		}
		cat.Children = &catChildren
		saveChannels[cat.AppId] = cat
	}
	guildRecord.Channels = saveChannels

	result := gormDB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "channels"}),
	}).Create(&guildRecord)
	if result.Error != nil {
		err = result.Error
		return errors.Wrap(result.Error, "Could not update guild record")
	}

	return nil
}
