package parse_test

import (
	"amizone/amizone/internal/mock"
	"amizone/amizone/internal/models"
	"amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestExaminationSchedule(t *testing.T) {
	testCases := []struct {
		name            string
		bodyFile        string
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

			fileReader, err := mock.FS.Open(testCase.bodyFile)
			g.Expect(err).ToNot(HaveOccurred())

			schedule, err := parse.ExaminationSchedule(fileReader)
			testCase.scheduleMatcher(g, schedule)
			testCase.errorMatcher(g, err)
		})
	}
}
