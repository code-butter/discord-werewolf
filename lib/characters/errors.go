package characters

import "fmt"

type TeamExhaustedError struct {
	team string
}

func (e *TeamExhaustedError) Error() string {
	return fmt.Sprintf("character pool for team '%s' is exhausted", e.team)
}

type ClassExhaustedError struct {
	team  string
	class string
}

func (e *ClassExhaustedError) Error() string {
	return fmt.Sprintf("character pool for class %s on team %s exhausted", e.class, e.team)
}

type SkipError struct {
}

func (e *SkipError) Error() string {
	return "no suitable characters for now, try again later"
}
