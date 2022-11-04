package amizone

import "github.com/ditsuke/go-amizone/amizone/internal/models"

// AttendanceRecord is a model for representing attendance record for a single course from the portal.
type AttendanceRecord models.AttendanceRecord

// AttendanceRecords is a model for representing attendance from the portal.
type AttendanceRecords models.AttendanceRecords

// ClassSchedule is a model for representing class schedule from the portal.
type ClassSchedule models.ClassSchedule

// ExamSchedule is a model for representing exam schedule from the portal.
type ExamSchedule models.ExaminationSchedule

// SemesterList is a model for representing semesters. Often, this model will be used
// for ongoing and past semesters for which information can be retrieved.
type SemesterList models.SemesterList

// CourseRef is a model for representing a minimal reference to a course, usually embedded in other models.
type CourseRef models.CourseRef

// Courses is a model for representing a list of courses from the portal. This model
// should most often be used to hold all courses for a certain semester.
type Courses models.Courses

// Marks is a model for representing marks (have/max).
type Marks models.Marks
