package parse_test

import (
	"amizone/internal/mock"
	"amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestLoggedIn(t *testing.T) {
	testcases := []struct {
		name     string
		bodyFile string
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

			fileReader, err := mock.FS.Open(tc.bodyFile)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(parse.LoggedIn(fileReader)).To(Equal(tc.expected))
		})
	}
}
