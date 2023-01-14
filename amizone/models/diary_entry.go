package models

// AmizoneDiaryEvent is the JSON format we expect from the Amizone diary events endpoint.
type AmizoneDiaryEvent struct {
	Type       string `json:"sType"` // "C" for course, "E" for event, "H" for holiday
	CourseName string `json:"title"`
	CourseCode string `json:"CourseCode"`
	Faculty    string `json:"FacultyName"`
	Room       string `json:"RoomNo"`
	Start      string `json:"start"` // Start and end keys are in the format "YYYY-MM-DD HH:MM:SS"
	End        string `json:"end"`
}

type AmizoneDiaryEvents []AmizoneDiaryEvent
