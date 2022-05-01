package models

type CourseAttendance struct {
	Course          *Course
	ClassesHeld     string
	ClassesAttended string
}

// AttendanceRecord maps course codes to courseAttendance structs
type AttendanceRecord map[string]*CourseAttendance
