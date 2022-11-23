package mock

import (
	"embed"
	"io/fs"
)

// filesystem is a mock filesystem with some files that can be used for testing.
//
//go:embed testdata
var filesystem embed.FS

type File string

// Open returns a fs.File interface to the file in filesystem, the mock filesystem.
func (f File) Open() (fs.File, error) {
	return filesystem.Open(string(f))
}

// Constants for file paths in the filesystem embedded filesystem.
const (
	DiaryEventsJSON     File = "testdata/diary_events.json"
	ExaminationSchedule File = "testdata/examination_schedule.html"
	HomePageLoggedIn    File = "testdata/home_page_logged_in.html"
	LoginPage           File = "testdata/login_page.html"
	CoursesPage         File = "testdata/my_courses.html"
	CoursesPageSemWise  File = "testdata/courses_semwise.html"
	IDCardPage          File = "testdata/id_card_page.html"
)
