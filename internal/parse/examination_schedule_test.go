package parse_test

import (
	"amizone/internal/models"
	"amizone/internal/parse"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestExaminationSchedule(t *testing.T) {
	testCases := []struct {
		name            string
		bodyFile        string
		scheduleMatcher func(g *GomegaWithT, schedule models.ExaminationSchedule)
		errorMatcher    func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid examination schedule page",
			bodyFile: ExaminationSchedulePage,
			scheduleMatcher: func(g *GomegaWithT, schedule models.ExaminationSchedule) {
				g.Expect(len(schedule)).To(Equal(8))
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "invalid examination schedule page",
			bodyFile: LoggedInHomePageFile,
			scheduleMatcher: func(g *GomegaWithT, schedule models.ExaminationSchedule) {
				g.Expect(schedule).To(BeEmpty())
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := os.Open(testCase.bodyFile)
			g.Expect(err).ToNot(HaveOccurred())

			schedule, err := parse.ExaminationSchedule(fileReader)
			testCase.scheduleMatcher(g, schedule)
			testCase.errorMatcher(g, err)
		})
	}
}
