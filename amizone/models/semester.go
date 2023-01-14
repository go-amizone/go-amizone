package models

// Semester models a semester reference on Amizone. We include both a semester "name" / label and a ref
// to decouple the way they're represented from their form values. These happen to be same at the time of
// modelling, however, so they might appear duplicitous.
type Semester struct {
	Name string
	Ref  string
}

// SemesterList is a model for representing semesters. Often, this model will be used
// for ongoing and past semesters for which information can be retrieved.
type SemesterList []Semester
