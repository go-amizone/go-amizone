package amizone

import "github.com/ditsuke/go-amizone/amizone/internal/models"

// Attendance is a model for representing attendance from the portal.
type Attendance models.AttendanceRecord

// ClassSchedule is a model for representing class schedule from the portal.
type ClassSchedule models.ClassSchedule

// ExamSchedule is a model for representing exam schedule from the portal.
type ExamSchedule models.ExaminationSchedule

// SemesterList is a model for representing semesters. Often, this model will be used
// for ongoing and past semesters for which information can be retrieved.
type SemesterList models.SemesterList

// Courses is a model for representing a list of courses from the portal. This model
// should most often be used to hold all courses for a certain semester.
type Courses models.Courses
