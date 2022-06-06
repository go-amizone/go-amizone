package mock

import "embed"

// FS is a mock filesystem with some files that can be used for testing.
//go:embed testdata
var FS embed.FS

// Constants for file paths in the FS embedded filesystem.
const (
	DiaryEventsJSON     = "testdata/diary_events.json"
	ExaminationSchedule = "testdata/examination_schedule.html"
	HomePageLoggedIn    = "testdata/home_page_logged_in.html"
	LoginPage           = "testdata/login_page.html"
	CoursesPage         = "testdata/my_courses.html"
)
