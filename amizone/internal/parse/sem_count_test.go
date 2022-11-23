package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
)

func TestSemesters(t *testing.T) {
	testCases := []struct {
		name             string
		bodyFile         mock.File
		semestersMatcher func(g *GomegaWithT, semesters models.SemesterList)
		errMatcher       func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid courses page",
			bodyFile: mock.CoursesPage,
			semestersMatcher: func(g *GomegaWithT, semesters models.SemesterList) {
				g.Expect(semesters).ToNot(BeNil())
				g.Expect(len(semesters)).To(Equal(4))
			},
			errMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "invalid courses page (login page)",
			bodyFile: mock.LoginPage,
			semestersMatcher: func(g *GomegaWithT, courses models.SemesterList) {
				g.Expect(courses).To(BeNil())
			},
			errMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err.Error()).To(ContainSubstring(parse.ErrFailedToParse))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			fileReader, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())
			semesters, err := parse.Semesters(fileReader)
			testCase.semestersMatcher(g, semesters)
			testCase.errMatcher(g, err)
		})
	}

}
