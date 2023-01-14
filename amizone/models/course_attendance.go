package models

type Attendance struct {
	ClassesHeld     int32
	ClassesAttended int32
}

// AttendanceRecord is a model for representing attendance record for a single course from the portal.
type AttendanceRecord struct {
	Attendance
	Course CourseRef
}

// AttendanceRecords is a model for representing attendance from the portal.
type AttendanceRecords []AttendanceRecord
