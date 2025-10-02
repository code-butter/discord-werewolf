package lib

func IsAdmin(i Interaction) (bool, error) {
	guild, err := i.Guild()
	if err != nil {
		return false, err
	}
	requester := i.Requester()
	if requester.ID == guild.OwnerID {
		return true, nil
	}
	return i.RequesterHasRole(RoleAdmin)
}
