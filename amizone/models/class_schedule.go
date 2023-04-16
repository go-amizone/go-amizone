package models

import (
	"sort"
	"time"

	"github.com/samber/lo"
)

type AttendanceState int

const (
	AttendanceStatePending AttendanceState = iota
	AttendanceStatePresent
	AttendanceStateAbsent
	AttendanceStateNA
	AttendanceStateInvalid
)

// ScheduledClass models the data extracted from the class schedule as found on the Amizone
// home page.
type ScheduledClass struct {
	Course    CourseRef
	StartTime time.Time
	EndTime   time.Time
	Faculty   string
	Room      string
	Attended  AttendanceState
}

// ClassSchedule is a model for representing class schedule from the portal.
type ClassSchedule []ScheduledClass

// Sort sorts the ClassSchedule by ScheduledClass.StartTime
func (s *ClassSchedule) Sort() {
	sort.Slice(*s, func(i, j int) bool {
		return (*s)[i].StartTime.Before((*s)[j].StartTime)
	})
}

func (s *ClassSchedule) FilterByDate(t time.Time) ClassSchedule {
	// Truncate the time to a day.
	targetDate := t.Truncate(time.Hour * 24)
	return lo.Filter(*s, func(class ScheduledClass, _ int) bool {
		timeDelta := class.StartTime.Sub(targetDate).Hours()
		return timeDelta > 0 && timeDelta < 24
	})
}
