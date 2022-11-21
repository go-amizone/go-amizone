package models

import (
	"strings"

	"k8s.io/klog/v2"
)

const (
	ColorAttendanceAbsent  = "#4FCC4F"
	ColorAttendancePending = "#3A87AD"
	ColorAttendancePresent = "#F00"
	ColorAttendanceNA      = ""
)

// AmizoneDiaryEvent is the JSON format we expect from the Amizone diary events endpoint.
type AmizoneDiaryEvent struct {
	Type            string `json:"sType"` // "C" for course, "E" for event, "H" for holiday
	CourseName      string `json:"title"`
	CourseCode      string `json:"CourseCode"`
	Faculty         string `json:"FacultyName"`
	Room            string `json:"RoomNo"`
	Start           string `json:"start"` // Start and end keys are in the format "YYYY-MM-DD HH:MM:SS"
	End             string `json:"end"`
	AttendanceColor string `json:"AttndColor"`
}

func (e *AmizoneDiaryEvent) AttendanceState() AttendanceState {
	switch strings.ToUpper(e.AttendanceColor) {
	case ColorAttendanceAbsent:
		return AttendanceStateAbsent
	case ColorAttendancePending:
		return AttendanceStatePending
	case ColorAttendancePresent:
		return AttendanceStatePresent
	case ColorAttendanceNA:
		return AttendanceStateNA
	}

	klog.Errorf("Unknown attendance color: %s", e.AttendanceColor)
	return AttendanceStateInvalid
}

type AmizoneDiaryEvents []AmizoneDiaryEvent
