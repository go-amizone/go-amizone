package models

// CourseRef models a reference to a course, most useful for models that have references but not complete course
// information.
type CourseRef struct {
	Code string
	Name string
}

// Course models the data found on the Amizone courses page.
type Course struct {
	CourseRef
	Type          string
	Attendance    Attendance
	InternalMarks Marks  // 0, 0 if not available
	SyllabusDoc   string // This is, really, a link. Most often broken and useless.
}

type Courses []Course
