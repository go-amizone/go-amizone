package amizone

type course struct {
	code string
	name string
}

type courseAttendance struct {
	course          *course
	classesHeld     string
	classesAttended string
}

// AttendanceRecord maps course codes to courseAttendance structs
type AttendanceRecord map[string]*courseAttendance
