package models

type CourseAttendance struct {
	Course          *Course
	ClassesHeld     int
	ClassesAttended int
}

// AttendanceRecord maps course codes to courseAttendance structs
type AttendanceRecord []CourseAttendance
