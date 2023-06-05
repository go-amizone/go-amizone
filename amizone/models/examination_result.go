package models

import (
	"time"
)

type ExamResultRecord struct {
	Course CourseRef
	CourseResult
}

// CourseResult is a model to represent the course wise result in the examinations result page
type CourseResult struct {
	Score       Score
	Credits     Credits
	PublishDate time.Time
}

type Score struct {
	Max        int
	Grade      string
	GradePoint int
}

type Credits struct {
	Acquired  int
	Effective int
	Points    int
}

// OverallResult is a model to represent the semester result, with the GPA etc in the examination result page
type OverallResult struct {
	Semester                    Semester
	SemesterGradePointAverage   float32
	CumulativeGradePointAverage float32
}

// ExamResultRecords includes the result for every course in an array and the
// overall result of every semester up to that point
type ExamResultRecords struct {
	CourseWise []ExamResultRecord
	Overall    []OverallResult
}
