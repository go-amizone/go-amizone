package models

import (
	"sort"
	"time"
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
	var filtered ClassSchedule
	for _, class := range *s {
		// Truncate the time to a day.
		tDate := t.Truncate(time.Hour * 24)

		if difference := class.StartTime.Sub(tDate).Hours(); difference > 0 && difference < 24 {
			filtered = append(filtered, class)
		}
	}
	return filtered
}
