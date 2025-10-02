package lib

var Roles map[string]WerewolfRole

type WerewolfRole struct {
	Name  string
	Color int
}

const RolePlaying = "Playing"
const RoleAlive = "Alive"
const RoleDead = "Dead"
const RoleAdmin = "Admin"

func init() {
	rolesArray := []WerewolfRole{
		{
			Name:  RolePlaying,
			Color: 0xFFDD81,
		},
		{
			Name:  RoleAlive,
			Color: 0x4ADC3D,
		},
		{
			Name:  RoleDead,
			Color: 0xBF0010,
		},
		{
			Name:  RoleAdmin,
			Color: 0x2025B7,
		},
	}
	Roles = map[string]WerewolfRole{}
	for _, role := range rolesArray {
		Roles[role.Name] = role
	}
}
