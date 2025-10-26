package lib

import "github.com/samber/do"

type SessionArgs struct {
	Session  DiscordSession
	Injector *do.Injector
}
