package mock

import "embed"

// FS is a mock filesystem with some files that can be used for testing.
//go:embed testdata
var FS embed.FS

// Constants for files in the FS filesystem.
const (
	DiaryEventsJSON     = "testdata/diary_events.json"
	ExaminationSchedule = "testdata/examination_schedule.html"
	HomePageLoggedIn    = "testdata/home_page_logged_in.html"
	LoginPage           = "testdata/login_page.html"
)
