package amizone

import (
	"github.com/ditsuke/go-amizone/amizone/internal/models"
)

// ClientInterface is an exported interface for amizoneClient to make mocking and testing more convenient.
type ClientInterface interface {
	DidLogin() bool
	GetAttendance() (models.AttendanceRecord, error)
	GetClassSchedule(date Date) (models.ClassSchedule, error)
	GetExamSchedule() (*models.ExaminationSchedule, error)
}
