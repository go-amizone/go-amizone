package parse_test

import (
	"amizone/amizone/internal/mock"
	"amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
	"testing"
)

func TestVerificationToken(t *testing.T) {
	testCases := []struct {
		name          string
		bodyFile      string
		expectedToken string
	}{
		{
			name:          "login page, verification token exists",
			bodyFile:      mock.LoginPage,
			expectedToken: "LV571ePb0TV-evRywWVGfbpe5PE71EpyM2U_9MGu69GA8-tlD4TaVd265sXZPoPyA2Xh2qV7D2t-8yKJWYzK17wyEMKuPseFtRk25WAqeC81",
		},
		{
			name:          "home page, verification token does not exist",
			bodyFile:      mock.HomePageLoggedIn,
			expectedToken: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			body, err := mock.FS.Open(tc.bodyFile)
			g.Expect(err).ToNot(HaveOccurred())

			token := parse.VerificationToken(body)
			g.Expect(token).To(Equal(tc.expectedToken))
		})
	}
}
