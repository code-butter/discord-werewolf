package characters

type DbCharacter struct {
	Id    int
	Label string
}

type Character struct {
	*DbCharacter
	*GameScore
	Team     string
	Class    string
	WinCount bool // does this member count toward the win conditions?
}

type ClassAssigner func(balancer GameBalance) (character Character, followupRequested bool, err error)

type GameScore struct {
	WolfScore    int
	VampireScore int
}

type GameBalance struct {
	*GameScore
	RandomMode    bool                // Ignore the score
	AcceptableIds []int               // Character IDs that can be used
	AssignedIds   map[int]int         // IDs already assigned and count
	SkipClasses   map[string][]string // Teams -> classes that need to be skipped
}
