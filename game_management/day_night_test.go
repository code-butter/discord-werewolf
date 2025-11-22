package game_management

import (
	"discord-werewolf/lib"
	"discord-werewolf/lib/testlib"
	"testing"
	"time"

	"github.com/samber/do"
)

func TestSystemAutoStarts(t *testing.T) {
	loopDelay := time.Millisecond * 10
	testDelay := time.Millisecond * 150

	clock := testlib.NewMockClock(time.Now().Add(-24 * time.Hour))
	clock.SetTime(19, 0, 0) // starts in the frozen state

	args := testlib.StartTestGame(10, 5, func(injector *do.Injector) {
		do.ProvideValue[lib.Clock](injector, clock)
	})
	guild, err := args.Session.Guild()
	if err != nil {
		t.Fatal(err)
	}
	guildId := guild.ID

	listeners := do.MustInvoke[*lib.GameListeners](args.Injector)
	settings := do.MustInvoke[*lib.GuildSettings](args.Injector)

	if err := settings.SetDayTime(guildId, "09:00"); err != nil {
		t.Fatal(err)
	}
	if err := settings.SetNightTime(guildId, "15:00"); err != nil {
		t.Fatal(err)
	}

	var dayInvoked, nightInvoked bool

	dayListener := func(*lib.SessionArgs, lib.DayStartData) error {
		dayInvoked = true
		return nil
	}

	nightListener := func(*lib.SessionArgs, lib.NightStartData) error {
		nightInvoked = true
		return nil
	}

	listeners.DayStart.Add(dayListener)
	listeners.NightStart.Add(nightListener)

	go TimedDayNight(args.Injector, loopDelay)

	// Ensure not getting called
	time.Sleep(testDelay)
	if dayInvoked || nightInvoked {
		t.Fatal("day or night was invoked before time")
	}

	// Check day time
	clock.Add(24 * time.Hour)
	clock.SetTime(9, 5, 0)
	time.Sleep(testDelay)
	if !dayInvoked {
		t.Fatal("day was not invoked when it should have been")
	}
	if nightInvoked {
		t.Fatal("night was invoked when it should not have been")
	}

	// Check night time
	dayInvoked = false
	clock.SetTime(19, 19, 0)
	time.Sleep(testDelay)
	if dayInvoked {
		t.Fatal("day was invoked when it should not have been")
	}
	if !nightInvoked {
		t.Fatal("night was not invoked when it should have been")
	}
}
