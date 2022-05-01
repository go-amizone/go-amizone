package parse_test

import (
	"GoFriday/internal/parse"
	. "github.com/onsi/gomega"
	"os"
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
			bodyFile: LoggedInHomePageFile,
			expected: true,
		},
		{
			name:     "not logged in",
			bodyFile: LoginPageFile,
			expected: false,
		},
		{
			name:     "json schedule",
			bodyFile: ScheduleJsonNonEmpty,
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			fileReader, err := os.Open(tc.bodyFile)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(parse.LoggedIn(fileReader)).To(Equal(tc.expected))
		})
	}
}
