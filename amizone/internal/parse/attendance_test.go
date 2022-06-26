package parse_test

import (
	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestAttendance(t *testing.T) {
	// @todo add test cases to cover more scenarios
	testCases := []struct {
		name              string
		bodyFile          mock.File
		attendanceMatcher func(g *GomegaWithT, attendance *models.AttendanceRecords)
		errorMatcher      func(g *GomegaWithT, err error)
	}{
		{
			name:     "logged in home page",
			bodyFile: mock.HomePageLoggedIn,
			attendanceMatcher: func(g *GomegaWithT, attendance *models.AttendanceRecords) {
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

			fileReader, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())

			attendance, err := parse.Attendance(fileReader)
			testCase.attendanceMatcher(g, &attendance)
			testCase.errorMatcher(g, err)
		})
	}
}
