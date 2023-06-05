package models

import (
	"time"
)

type ExamResultRecord struct {
	Course CourseRef
	Result CourseResult
}

// CourseResult is a model to represent the course wise result in the examinations result page
type CourseResult struct {
	MaxTotal             int
	AquiredCreditUnits   int
	GradeObtained        string
	GradePoint           int
	CreditPoints         int
	EffectiveCreditUnits int
	PublishDate          time.Time
}

// OverlallResult is a model to represent the semester result, with the GPA etc in the examination result page
type OverallResult struct {
	Semester                       Semester
	SemesterGradePointAverage      float32
	CummulatitiveGradePointAverage float32
}

// ExamResultRecords includes the result for every course in an array and the
// overall result of every semester upto that point
type ExamResultRecords struct {
	CourseWise []ExamResultRecord
	Overall    []OverallResult
}
