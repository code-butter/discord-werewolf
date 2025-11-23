package authorizors

import (
	"context"
	"discord-werewolf/lib"
	"discord-werewolf/lib/models"
	"slices"

	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func IsAdmin(ia *lib.InteractionArgs) error {
	guild, err := ia.Session.Guild()
	if err != nil {
		return err
	}
	requester := ia.Interaction.Requester()
	if requester.ID == guild.OwnerID {
		return nil
	}
	yes, err := ia.Interaction.RequesterHasRole(lib.RoleAdmin)
	if err != nil {
		return err
	}
	if !yes {
		return lib.NewPermissionDeniedError("")
	}
	return nil
}

func IsAlive(ia *lib.InteractionArgs) error {
	yes, err := ia.Interaction.RequesterHasRole(lib.RoleAlive)
	if err != nil {
		return err
	}
	if !yes {
		return lib.NewPermissionDeniedError("You're dead, bub.")
	}
	return nil
}

func IsDayTime(ia *lib.InteractionArgs) error {
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}
	if !guild.DayNight {
		return lib.NewPermissionDeniedError("It's night time. You can do this tomorrow.")
	}
	return nil
}

func IsNightTime(ia *lib.InteractionArgs) error {
	guild, err := ia.AppGuild()
	if err != nil {
		return err
	}
	if guild.DayNight {
		return lib.NewPermissionDeniedError("It's day time. You can do this tonight.")
	}
	return nil
}

func CharacterExists(option_name string) func(ia *lib.InteractionArgs) error {
	return func(ia *lib.InteractionArgs) error {
		db := do.MustInvoke[*gorm.DB](ia.Injector)
		ctx := do.MustInvoke[context.Context](ia.Injector)

		option := ia.Interaction.CommandData().GetOption(option_name)
		if option == nil {
			return nil // This handles scenarios where options are not required to be set
		}
		userId := option.Value.(string)
		_, err := gorm.G[models.GuildCharacter](db).
			Where("id = ?", userId).
			First(ctx)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return lib.NewPermissionDeniedError("They're not playing! Choose someone else.")
			} else {
				return err
			}
		}
		aliveRole, err := ia.Session.GetRoleByName(lib.RoleAlive)
		if err != nil {
			return err
		}
		member, err := ia.Session.GuildMember(userId)
		if err != nil {
			return err
		}
		if !slices.Contains(member.Roles, aliveRole.ID) {
			return lib.NewPermissionDeniedError("They're not on this mortal coil. Try choosing someone who is alive.")
		}
		return nil
	}
}
