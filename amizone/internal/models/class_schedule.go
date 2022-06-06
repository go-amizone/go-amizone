package models

import (
	"sort"
	"time"
)

type ScheduledClass struct {
	Course    CourseRef
	StartTime time.Time
	EndTime   time.Time
	Faculty   string
	Room      string
}

// ClassSchedule is an array of ScheduledClass, typically for a single day
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
