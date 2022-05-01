//go:build integration

package GoFriday_test

import (
	"GoFriday"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

func TestIntegrateNewClient(t *testing.T) {
	g := NewGomegaWithT(t)

	validUser := os.Getenv("AMIZONE_USERNAME")
	validPassword := os.Getenv("AMIZONE_PASSWORD")

	g.Expect(validUser).ToNot(BeEmpty(), "AMIZONE_USERNAME environment variable is not set")
	g.Expect(validPassword).ToNot(BeEmpty(), "AMIZONE_PASSWORD environment variable is not set")

	testCases := []struct {
		name          string
		credentials   GoFriday.Credentials
		errorMatcher  func(g *GomegaWithT, err error)
		clientMatcher func(g *GomegaWithT, client GoFriday.ClientInterface)
	}{
		{
			name:        "valid credentials",
			credentials: GoFriday.Credentials{Username: validUser, Password: validPassword},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(BeNil())
			},
			clientMatcher: func(g *GomegaWithT, client GoFriday.ClientInterface) {
				g.Expect(client).ToNot(BeNil())
				g.Expect(client.DidLogin()).To(BeTrue())
			},
		},
		{
			name:        "invalid credentials",
			credentials: GoFriday.Credentials{Username: "this-user-does-not-exist", Password: "neither-does-this-password"},
			errorMatcher: func(g *GomegaWithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(GoFriday.ErrFailedLogin))
			},
			clientMatcher: func(g *GomegaWithT, client GoFriday.ClientInterface) {
				g.Expect(client).ToNot(BeNil())
				g.Expect(client.DidLogin()).To(BeFalse())
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			client, err := GoFriday.NewClient(tc.credentials, nil)
			tc.errorMatcher(g, err)
			tc.clientMatcher(g, client)
		})
	}
}
