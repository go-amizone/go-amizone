package mock

import "embed"

//go:embed testdata
var FS embed.FS

const (
	DiaryEventsJSON     = "testdata/diary_events.json"
	ExaminationSchedule = "testdata/examination_schedule.html"
	HomePageLoggedIn    = "testdata/home_page_logged_in.html"
	LoginPage           = "testdata/login_page.html"
)
