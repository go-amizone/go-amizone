package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
)

func TestProfile(t *testing.T) {
	testCases := []struct {
		name           string
		bodyFile       mock.File
		profileMatcher func(g *GomegaWithT, profile *models.Profile)
		errMatcher     func(g *GomegaWithT, err error)
	}{
		{
			name:     "valid profile page",
			bodyFile: mock.IDCardPage,
			profileMatcher: func(g *GomegaWithT, profile *models.Profile) {
				g.Expect(profile).ToNot(BeNil())
			},
			errMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:     "login page",
			bodyFile: mock.LoginPage,
			profileMatcher: func(g *GomegaWithT, profile *models.Profile) {
				g.Expect(profile).To(BeNil())
			},
			errMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(parse.ErrFailedToParse))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			body, err := testCase.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())
			profile, err := parse.Profile(body)
			testCase.profileMatcher(g, profile)
			testCase.errMatcher(g, err)
		})
	}
}
