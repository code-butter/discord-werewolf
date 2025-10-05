package lib

func IsAdmin(ia *InteractionArgs) (bool, error) {
	guild, err := ia.Session.Guild()
	if err != nil {
		return false, err
	}
	requester := ia.Interaction.Requester()
	if requester.ID == guild.OwnerID {
		return true, nil
	}
	return ia.Interaction.RequesterHasRole(RoleAdmin)
}
