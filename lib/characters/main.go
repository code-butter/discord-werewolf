package characters

const TeamVillager = "villagers"
const TeamWerewolf = "werewolves"
const TeamVampire = "vampires"
const TeamChaos = "chaos"

const Werewolf = "werewolf"
const WerewolfCub = "werewolf-cub"
const Villager = "villager"
const KingVampire = "king-vampire"
const Witch = "witch"
const Bodyguard = "bodyguard"
const Seer = "seer"
const Fool = "fool"
const ApprenticeSeer = "apprentice-seer"
const Mason = "mason"
const Baker = "baker"
const Hunter = "hunter"
const Monarch = "monarch"
const Lycan = "lycan"
const Granny = "granny"
const Mutated = "mutated"
const ChaosDemon = "chaos-demon"
const Cupid = "cupid"

const SecondaryVampire = "vampire"
const SecondaryHenchman = "henchman"

var Teams = map[string]Team{
	TeamWerewolf: {
		Id:    TeamWerewolf,
		Label: "Werewolves",
	},
	TeamVillager: {
		Id:    TeamVillager,
		Label: "Villagers",
	},
	TeamVampire: {
		Id:    TeamVampire,
		Label: "Vampires",
	},
	TeamChaos: {
		Id:    TeamChaos,
		Label: "Chaos",
	},
}

var FakeVillagerIds = []string{
	Baker, Hunter, Villager, ApprenticeSeer,
}

var Characters = map[string]Character{
	Villager: {
		Id:          Villager,
		TeamId:      TeamVillager,
		Label:       "Villager",
		GameScore:   1,
		Description: "A simple, common villager.",
		WinCount:    true,
	},
	Werewolf: {
		Id:          Werewolf,
		TeamId:      TeamWerewolf,
		GameScore:   -3,
		WinCount:    true,
		Label:       "Werewolf",
		Description: "Each night a wolf can vote on who to kill as a pack.",
	},
	WerewolfCub: {
		Id:          WerewolfCub,
		TeamId:      TeamWerewolf,
		GameScore:   -4,
		WinCount:    true,
		FakeIds:     []string{WerewolfCub},
		Label:       "Werewolf Cub",
		Description: "When the cub dies, the werewolf pack becomes enraged and is allowed to kill two at night. Appears as a wolf.",
	},
	Witch: {
		Id:          Witch,
		TeamId:      TeamWerewolf,
		GameScore:   -3,
		WinCount:    false,
		Label:       "Witch",
		Description: "Each night the witch curses a villager. When the witch dies by hanging or shooting all the cursed die.",
	},
	KingVampire: {
		Id:          KingVampire,
		TeamId:      TeamVampire,
		GameScore:   -1,
		WinCount:    true,
		Label:       "King Vampire",
		Description: "Vampire king bites players to turn them into vampires. Kings are optionally protected when biting a wolf or its victim and can optionally turn a villager on the first bite.",
	},
	Bodyguard: {
		Id:          Bodyguard,
		TeamId:      TeamVillager,
		GameScore:   3,
		WinCount:    true,
		Label:       "Bodyguard",
		Description: "Protects villagers from wolves and vampires at night. Is notified if guarding or protecting from a vampire.",
	},
	Seer: {
		Id:          Seer,
		TeamId:      TeamVillager,
		GameScore:   3,
		WinCount:    true,
		Label:       "Seer",
		Description: "Can divine the if someone is a wolf. Given own channel to investigate.",
	},
	Fool: {
		Id:          Fool,
		TeamId:      TeamVillager,
		GameScore:   -2,
		WinCount:    true,
		Label:       "Fool",
		Description: "Appears as a seer and given own channel. Results are reported randomly (75% villager, 25% wolf).",
		FakeIds:     []string{Seer},
	},
	ApprenticeSeer: {
		Id:          ApprenticeSeer,
		TeamId:      TeamVillager,
		GameScore:   2,
		WinCount:    true,
		Label:       "Apprentice Seer",
		Description: "Becomes the seer upon the seer's death.",
	},
	Mason: {
		Id:          Mason,
		TeamId:      TeamVillager,
		GameScore:   2,
		WinCount:    true,
		Label:       "Mason",
		Description: "A member of a secret club. Optionally bodyguard and seer may join the mason channel.",
	},
	Baker: {
		Id:          Baker,
		TeamId:      TeamVillager,
		GameScore:   -2,
		WinCount:    true,
		Label:       "Baker",
		Description: "After the death of the baker, a villager/witch dies of starvation each morning.",
	},
	Hunter: {
		Id:          Hunter,
		TeamId:      TeamVillager,
		GameScore:   1,
		WinCount:    true,
		Label:       "Mason",
		Description: "Shoots another player upon death. If no choice is made in time a random player is shot.",
	},
	Monarch: {
		Id:          Monarch,
		TeamId:      TeamVillager,
		GameScore:   2,
		WinCount:    true,
		Label:       "Monarch",
		Description: "Grants power gifts to other players at nights.",
	},
	Lycan: {
		Id:          Lycan,
		TeamId:      TeamVillager,
		GameScore:   -1,
		WinCount:    true,
		FakeIds:     FakeVillagerIds,
		Label:       "Lycan",
		Description: "Villager who appears as a wolf to the seer. Can optionally be protected from wolf attacks.",
	},
	Granny: {
		Id:          Granny,
		TeamId:      TeamVillager,
		GameScore:   1,
		WinCount:    true,
		Label:       "Grouchy Granny",
		Description: "Chooses one person each day to mute from all other channels and spend time in her channel.",
	},
	Mutated: {
		Id:          Mutated,
		TeamId:      TeamVillager,
		GameScore:   -2,
		WinCount:    true,
		FakeIds:     FakeVillagerIds,
		Label:       "Mutated Villager",
		Description: "Mutated villager is told a villager role, but if attacked by wolves or vampires mutates to be on that team.",
	},
	ChaosDemon: {
		Id:          ChaosDemon,
		TeamId:      TeamChaos,
		GameScore:   -3,
		Label:       "Chaos Demon",
		Description: "Chooses a victim to hang. Only wins if the victim dies by hanging, dies if the victim dies any other way.",
	},
	Cupid: {
		Id:          Cupid,
		TeamId:      TeamChaos,
		Label:       "Cupid",
		GameScore:   0,
		Description: "Chooses a couple to be in love. If the lovers are the last up to or including Cupid, they all win.",
	},
}

var SecondaryCharacters = map[string]SecondaryCharacter{
	SecondaryVampire: {
		Id:          SecondaryVampire,
		TeamId:      TeamVampire,
		Label:       "Vampire",
		Description: "Vampires are created by other vampires and join the team. They can attack and turn villagers at night. Two bites turns a villager, but not a wolf nor the chaos demon. Regular vampires may die if biting a wolf or attacking the wolves' victim.",
		WinCount:    true,
	},
	SecondaryHenchman: {
		Id:          SecondaryHenchman,
		TeamId:      TeamWerewolf,
		Label:       "Henchman",
		Description: "Optionally, a henchman is created on the first night instead of a villager dying on the first night of gameplay.",
		WinCount:    false,
	},
}

func init() {
	for _, c := range Characters {
		team := Teams[c.TeamId]
		team.Characters = append(team.Characters, c)
		Teams[c.TeamId] = team
	}
}
