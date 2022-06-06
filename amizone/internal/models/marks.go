package models

type Marks struct {
	Have float32
	Max  float32
}

// Available indicates if marks were available
func (m Marks) Available() bool {
	return m.Max != 0
}
