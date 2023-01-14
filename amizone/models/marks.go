package models

// Marks is a model for representing marks (have/max).
type Marks struct {
	Have float32
	Max  float32
}

// Available indicates if marks were available
func (m Marks) Available() bool {
	return m.Max != 0
}
