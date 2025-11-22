package lib

import "time"

type Clock interface {
	Now() time.Time
	Sleep(time.Duration)
	Since(time.Time) time.Duration
}

type RealClock struct {
}

func (r RealClock) Now() time.Time {
	return time.Now()
}

func (r RealClock) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func (r RealClock) Since(t time.Time) time.Duration {
	return time.Since(t)
}
