package models

// Semester models a semester reference on Amizone. We include both a semester "name" / label and a ref
// to decouple the way they're represented from their form values. These happen to be same at the time of
// modelling, however, so they might appear duplicitous.
type Semester struct {
	Name string
	Ref  string
}

type SemesterList []Semester
