package models

import (
	"database/sql/driver"
	"time"
)

type TimeOnly struct {
	*time.Time
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

func (m TimeOnly) BeforeOrOn(t time.Time) bool {
	mTime := []int{m.Time.Hour(), m.Time.Minute(), m.Time.Second()}
	tTime := []int{t.Hour(), t.Minute(), t.Second()}
	for i, mt := range mTime {
		if mt < tTime[i] {
			return true
		}
		if mt > tTime[i] {
			return false
		}
	}
	return true
}

func (m TimeOnly) AfterOrOn(t time.Time) bool {
	mTime := []int{m.Time.Hour(), m.Time.Minute(), m.Time.Second()}
	tTime := []int{t.Hour(), t.Minute(), t.Second()}
	for i, mt := range mTime {
		if mt > tTime[i] {
			return true
		}
		if mt < tTime[i] {
			return false
		}
	}
	return true
}
