package models

import "time"

type ScheduledExam struct {
	Course CourseRef
	Time   time.Time
	Mode   string
}

// ExaminationSchedule is a model for representing exam schedule from the portal.
type ExaminationSchedule struct {
	Title string
	Exams []ScheduledExam
}
