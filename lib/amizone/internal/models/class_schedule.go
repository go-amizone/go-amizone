package models

import (
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
