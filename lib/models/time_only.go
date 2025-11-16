package models

import (
	"database/sql/driver"
	"time"
)

type TimeOnly struct {
	*time.Time
}

func NewTimeOnly(hours, minutes, seconds int) *TimeOnly {
	time := time.Date(0, 0, 0, hours, minutes, seconds, 0, time.UTC)
	return &TimeOnly{Time: &time}
}

func (m TimeOnly) Value() (driver.Value, error) {
	return m.Time.Format(time.TimeOnly), nil
}
func (m *TimeOnly) Scan(value interface{}) error {
	t, err := time.Parse(time.TimeOnly, value.(string))
	if err != nil {
		return err
	}
	m.Time = &t
	return nil
}

func (m TimeOnly) GormDataType() string {
	return "time_only"
}

func (m TimeOnly) TimeOnDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), m.Hour(), m.Minute(), m.Second(), 0, t.Location())
}
