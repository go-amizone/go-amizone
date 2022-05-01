package parse_test

import (
	"amizone/internal/mock"
	"amizone/internal/models"
	"amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAttendance(t *testing.T) {
	// @todo add test cases to cover more scenarios
	testCases := []struct {
		name              string
		bodyFile          string
		attendanceMatcher func(g *GomegaWithT, attendance *models.AttendanceRecord)
		errorMatcher      func(g *GomegaWithT, err error)
	}{
		{
			name:     "logged in home page",
			bodyFile: mock.HomePageLoggedIn,
			attendanceMatcher: func(g *GomegaWithT, attendance *models.AttendanceRecord) {
				g.Expect(len(*attendance)).To(Equal(8))
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := mock.FS.Open(testCase.bodyFile)
			g.Expect(err).ToNot(HaveOccurred())

			attendance, err := parse.Attendance(fileReader)
			testCase.attendanceMatcher(g, &attendance)
			testCase.errorMatcher(g, err)
		})
	}
}
