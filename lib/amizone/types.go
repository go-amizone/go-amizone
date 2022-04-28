package amizone

type course struct {
	code string
	name string
}

type courseAttendance struct {
	classesHeld     string
	classesAttended string
}

type attendanceRecord map[*course]*courseAttendance
