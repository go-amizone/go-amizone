package response_models

import "github.com/ditsuke/go-amizone/amizone"

// Attendance is the response model for the attendance endpoint
type Attendance amizone.Attendance

// ClassSchedule is the model for the class schedule endpoint
type ClassSchedule amizone.ClassSchedule

// ExamSchedule is the model for the exam schedule endpoint
type ExamSchedule amizone.ExamSchedule

type errorMessage struct {
	Message string
}

// ErrorResponse is the model used when things go wrong
type ErrorResponse []errorMessage
