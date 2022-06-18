package parse_test

import (
	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestExaminationSchedule(t *testing.T) {
	testCases := []struct {
		name            string
		bodyFile        mock.File
		scheduleMatcher func(g *GomegaWithT, schedule *models.ExaminationSchedule)
		errorMatcher    func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid examination schedule page",
			bodyFile: mock.ExaminationSchedule,
			scheduleMatcher: func(g *GomegaWithT, schedule *models.ExaminationSchedule) {
				g.Expect(len(schedule.Exams)).To(Equal(8))
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "invalid examination schedule page",
			bodyFile: mock.HomePageLoggedIn,
			scheduleMatcher: func(g *GomegaWithT, schedule *models.ExaminationSchedule) {
				g.Expect(schedule).To(BeNil())
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())

			schedule, err := parse.ExaminationSchedule(fileReader)
			testCase.scheduleMatcher(g, schedule)
			testCase.errorMatcher(g, err)
		})
	}
}
