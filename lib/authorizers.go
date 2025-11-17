package lib

func IsAdmin(ia *InteractionArgs) error {
	guild, err := ia.Session.Guild()
	if err != nil {
		return err
	}
	requester := ia.Interaction.Requester()
	if requester.ID == guild.OwnerID {
		return nil
	}
	yes, err := ia.Interaction.RequesterHasRole(RoleAdmin)
	if err != nil {
		return err
	}
	if !yes {
		return NewPermissionDeniedError("")
	}
	return nil
}

func IsAlive(ia *InteractionArgs) error {
	yes, err := ia.Interaction.RequesterHasRole(RoleAlive)
	if err != nil {
		return err
	}
	if !yes {
		return NewPermissionDeniedError("You're dead, bub.")
	}
	return nil
}
