package mock

import (
	"embed"
	"io/fs"
)

// fileSystem is a mock filesystem with some files that can be used for testing.
//go:embed testdata
var fileSystem embed.FS

type File string

// Open returns a fs.File interface to the file in fileSystem, the mock filesystem.
func (f File) Open() (fs.File, error) {
	return fileSystem.Open(string(f))
}

// Constants for file paths in the fileSystem embedded filesystem.
const (
	DiaryEventsJSON     File = "testdata/diary_events.json"
	ExaminationSchedule File = "testdata/examination_schedule.html"
	HomePageLoggedIn    File = "testdata/home_page_logged_in.html"
	LoginPage           File = "testdata/login_page.html"
	CoursesPage         File = "testdata/my_courses.html"
)
