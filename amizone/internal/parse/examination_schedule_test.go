package parse_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/models"
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
		{
			name:     "examination schedule with location",
			bodyFile: mock.ExaminationScheduleWithLocation,
			scheduleMatcher: func(g *GomegaWithT, schedule *models.ExaminationSchedule) {
				expected := ReadExpectedFile(mock.ExpectedExamScheduleWithRoom, g)
				g.Expect(toJSON(schedule, g)).To(MatchJSON(expected))
			},
			errorMatcher: func(g *GomegaWithT, err error) {

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
