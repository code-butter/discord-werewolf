package testlib

import "time"

type MockClock struct {
	currentTime time.Time
	advanceFrom *time.Time
}

func NewMockClock(currentTime time.Time) *MockClock {
	return &MockClock{currentTime: currentTime}
}

func (m *MockClock) Now() time.Time {
	if m.advanceFrom == nil {
		return m.currentTime
	}
	return m.currentTime.Add(time.Since(*m.advanceFrom))
}

func (m *MockClock) Sleep(duration time.Duration) {
	sleepEnd := m.Now().Add(duration)
	for m.Now().Before(sleepEnd) {
		time.Sleep(10 * time.Millisecond)
	}
}

func (m *MockClock) Since(t time.Time) time.Duration {
	return m.Now().Sub(t)
}

func (m *MockClock) Set(newTime time.Time) {
	m.currentTime = newTime
}

func (m *MockClock) SetTime(hour int, minute int, seconds int) {
	now := m.Now()
	m.Set(time.Date(now.Year(), now.Month(), now.Day(), hour, minute, seconds, 0, now.Location()))
}

func (m *MockClock) SetDate(year int, month time.Month, day int) {
	now := m.Now()
	m.Set(time.Date(year, month, day, now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), now.Location()))
}

func (m *MockClock) SetLocation(loc *time.Location) {
	m.Set(m.Now().In(loc))
}

func (m *MockClock) Freeze() {
	if m.advanceFrom != nil {
		m.currentTime = m.currentTime.Add(time.Since(*m.advanceFrom))
	}
	m.advanceFrom = nil
}

func (m *MockClock) Unfreeze() {
	now := time.Now()
	m.advanceFrom = &now
}

func (m *MockClock) Add(duration time.Duration) {
	m.currentTime = m.currentTime.Add(duration)
}
