package models

import "time"

type ScheduledExam struct {
	Course *Course
	Time   time.Time
	Mode   string
}

type ExaminationSchedule []*ScheduledExam
