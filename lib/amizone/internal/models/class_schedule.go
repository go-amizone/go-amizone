package models

import (
	"sort"
	"time"
)

type ScheduledClass struct {
	Course    *Course
	StartTime time.Time
	EndTime   time.Time
	Faculty   string
	Room      string
}

// ClassSchedule is an array of ScheduledClass, typically for a single day
type ClassSchedule []*ScheduledClass

// Sort sorts the ClassSchedule by ScheduledClass.StartTime
func (s *ClassSchedule) Sort() {
	sort.Slice(*s, func(i, j int) bool {
		return (*s)[i].StartTime.Before((*s)[j].StartTime)
	})
}
