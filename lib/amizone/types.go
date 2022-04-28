package amizone

import (
	"time"
)

type course struct {
	code string
	name string
}

type courseAttendance struct {
	course          *course
	classesHeld     string
	classesAttended string
}

type Date struct {
	Year  int
	Month int
	Day   int
}

type scheduledClass struct {
	course    *course
	startTime time.Time
	endTime   time.Time
	faculty   string
	room      string
}

// AttendanceRecord maps course codes to courseAttendance structs
type AttendanceRecord map[string]*courseAttendance

// classSchedule is an array of scheduledClass, typically for a single day
type classSchedule []*scheduledClass
