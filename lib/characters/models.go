package characters

type Character struct {
	Id          string
	TeamId      string
	Label       string
	GameScore   int
	Class       string
	Description string
	WinCount    bool     // does this member count toward the win conditions?
	FakeIds     []string // Character IDs that this user could be reported as (for fool, cub, etc.)
}

type SecondaryCharacter struct {
	Id          string
	TeamId      string
	Label       string
	Description string
	WinCount    bool
}

type Team struct {
	Id    string
	Label string
}
