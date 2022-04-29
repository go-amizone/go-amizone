package amizone

// ClientInterface is an exported interface for amizoneClient to make mocking and testing more convenient.
type ClientInterface interface {
	GetAttendance() (AttendanceRecord, error)
	GetClassSchedule(date Date) (classSchedule, error)
}
