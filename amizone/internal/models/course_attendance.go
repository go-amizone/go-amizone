package models

type Attendance struct {
	ClassesHeld     int32
	ClassesAttended int32
}
type AttendanceRecord struct {
	Attendance
	Course CourseRef
}

// AttendanceRecords maps course codes to courseAttendance structs
type AttendanceRecords []AttendanceRecord
