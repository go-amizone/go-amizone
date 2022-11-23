package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
)

func TestLoggedIn(t *testing.T) {
	testcases := []struct {
		name     string
		bodyFile mock.File
		expected bool
	}{
		{
			name:     "logged in",
			bodyFile: mock.HomePageLoggedIn,
			expected: true,
		},
		{
			name:     "not logged in",
			bodyFile: mock.LoginPage,
			expected: false,
		},
		{
			name:     "json schedule",
			bodyFile: mock.DiaryEventsJSON,
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := tc.bodyFile.Open()

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(parse.LoggedIn(fileReader)).To(Equal(tc.expected))
		})
	}
}
