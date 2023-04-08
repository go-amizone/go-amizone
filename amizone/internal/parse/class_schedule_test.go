package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/models"
	. "github.com/onsi/gomega"
)

func TestClassSchedule(t *testing.T) {
	testCases := []struct {
		name            string
		bodyFile        mock.File
		scheduleMatcher func(g *GomegaWithT, schedule models.ClassSchedule)
		errorMatcher    func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid diary events json",
			bodyFile: mock.DiaryEventsJSON,
			scheduleMatcher: func(g *WithT, schedule models.ClassSchedule) {
				g.Expect(schedule).ToNot(BeNil())
				g.Expect(schedule).To(HaveLen(10))
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "invalid diary events json",
			bodyFile: mock.LoginPage,
			scheduleMatcher: func(g *GomegaWithT, schedule models.ClassSchedule) {
				g.Expect(schedule).To(BeNil())
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring("JSON decode"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())

			schedule, err := parse.ClassSchedule(fileReader)
			testCase.scheduleMatcher(g, schedule)
			testCase.errorMatcher(g, err)
		})
	}
}
