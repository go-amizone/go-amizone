package parse_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/models"
)

func TestExaminationResult(t *testing.T) {
	testCases := []struct {
		name          string
		bodyFile      mock.File
		resultMatcher func(g *GomegaWithT, result *models.ExamResultRecords)
		errorMatcher  func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid examination result page",
			bodyFile: mock.ExaminationResultPage,
			resultMatcher: func(g *GomegaWithT, result *models.ExamResultRecords) {
				g.Expect(len(result.CourseWise)).To(Equal(8))
				g.Expect(len(result.Overall)).To(Equal(3))
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "invalid examination result page",
			bodyFile: mock.HomePageLoggedIn,
			resultMatcher: func(g *GomegaWithT, result *models.ExamResultRecords) {
				g.Expect(result).To(BeNil())
			},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)

			fileReader, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())

			result, err := parse.ExaminationResult(fileReader)
			testCase.resultMatcher(g, result)
			testCase.errorMatcher(g, err)
		})
	}
}
