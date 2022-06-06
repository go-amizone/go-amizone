package models

type Attendance struct {
	ClassesHeld     int
	ClassesAttended int
}
type CourseAttendance struct {
	Attendance
	Course CourseRef
}

// AttendanceRecord maps course codes to courseAttendance structs
type AttendanceRecord []CourseAttendance
