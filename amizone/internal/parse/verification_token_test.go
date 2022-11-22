package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
)

func TestVerificationToken(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	testCases := []struct {
		name          string
		bodyFile      mock.File
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

			body, err := tc.bodyFile.Open()
			g.Expect(err).ToNot(HaveOccurred())

			token := parse.VerificationToken(body)
			g.Expect(token).To(Equal(tc.expectedToken))
		})
	}
}
