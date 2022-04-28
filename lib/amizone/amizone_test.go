package amizone_test

import (
	"GoFriday/lib/amizone"
	"errors"
	"fmt"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"testing"
)

const (
	fakeAmizoneUsername = "fakeUsername"
	fakeAmizonePassword = "fakePassword"

	fakeAuthCookie               = "fakeAuthCookie"
	fakeRequestVerificationToken = "fakeRequestVerificationToken"
	fakeSessionId                = "fakeSessionId"
)

type amizoneClientInterface interface {
	GetAttendance() (amizone.AttendanceRecord, error)
}

// @todo: implement test cases to test behavior when:
// - Amizone is not reachable
// - Amizone is reachable but login fails (invalid credentials, etc?)
func TestNewClient(t *testing.T) {
	g := NewGomegaWithT(t)

	defer gock.Off()

	err := gockRegisterLoginPage()
	gockRegisterLoginRequest(fakeAmizoneUsername, fakeAmizonePassword)

	g.Expect(err).ToNot(HaveOccurred(), "Failed to open mock login page")

	jar, err := cookiejar.New(nil)
	g.Expect(err).ToNot(HaveOccurred(), "Failed to create cookie jar")

	httpClient := &http.Client{Jar: jar}
	gock.InterceptClient(httpClient)

	c := amizone.Credentials{
		Username: fakeAmizoneUsername,
		Password: fakeAmizonePassword,
	}

	client, err := amizone.NewClient(c, httpClient)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(client).ToNot(BeNil())
}

// What are your expectations of this function?
// Login? No. That's not its responsibility.
// What we do expect is:
// It makes a request as the amizone client mocked would
// And then it retrieves the attendance record from the test page as it exists.
// Cases: Right record with the login mocked, no record with no login.
func TestAmizoneClient_GetAttendance(t *testing.T) {
	g := NewGomegaWithT(t)

	gock.DisableNetworking()
	defer gock.EnableNetworking()
	defer gock.Off()

	//Hacky initialization of the clients because we cannot refer to unexported types directly.
	nonLoggedInClient, err := amizone.NewClient(amizone.Credentials{}, nil)
	loggedInClient := nonLoggedInClient
	g.Expect(err).To(HaveOccurred(), "The amizone client shouldn't be logged in.")

	// Setup the logged-in and non logged-in amizone clients.
	func() {
		defer gock.Off()
		err := gockRegisterLoginPage()
		g.Expect(err).ToNot(HaveOccurred(), "Failed to register mock login page")
		gockRegisterLoginRequest(fakeAmizoneUsername, fakeAmizonePassword)
		loggedInClient, err = amizone.NewClient(amizone.Credentials{
			Username: fakeAmizoneUsername,
			Password: fakeAmizonePassword,
		}, nil)
		g.Expect(err).ToNot(HaveOccurred(), "set up mocked logged-in amizone client")
	}()

	testCases := []struct {
		name              string
		amizoneClient     amizoneClientInterface
		setup             func(g *WithT)
		attendanceMatcher func(g *WithT, attendance amizone.AttendanceRecord)
		errorMatcher      func(g *WithT, err error)
	}{
		{
			name:          "Logged in, expecting retrieval",
			amizoneClient: loggedInClient,
			setup: func(g *WithT) {
				err := gockRegisterHomePageLoggedIn()
				g.Expect(err).ToNot(HaveOccurred())
			},
			attendanceMatcher: func(g *WithT, attendance amizone.AttendanceRecord) {
				g.Expect(len(attendance)).To(Equal(8))
			},
			errorMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:          "Not logged in, expecting no retrieval",
			amizoneClient: nonLoggedInClient,
			setup: func(g *WithT) {
				err := gockRegisterHomePageLoggedIn()
				g.Expect(err).ToNot(HaveOccurred())
				gock.New("https://s.amizone.net").
					Get("/Home").
					Reply(http.StatusOK).
					BodyString("<html><body>Forbidden -- No Records for you</body></html>")
			},
			attendanceMatcher: func(g *WithT, attendance amizone.AttendanceRecord) {
				g.Expect(attendance).To(BeEmpty())
			},
			errorMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(Equal(amizone.ErrFailedAttendanceRetrieval))
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			defer gock.Off()

			c.setup(g)

			attendance, err := c.amizoneClient.GetAttendance()
			c.attendanceMatcher(g, attendance)
			c.errorMatcher(g, err)
		})
	}
}

func gockRegisterLoginPage() error {
	mockLogin, err := os.Open("testdata/login_page.html")
	if err != nil {
		return errors.New("Failed to open mock login page: " + err.Error())
	}

	gock.New("https://s.amizone.net").
		Get("/").
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockLogin)

	return nil
}

func gockRegisterLoginRequest(validUsername string, validPassword string) {
	gock.New("https://s.amizone.net").
		Post("/").
		MatchType("application/x-www-form-urlencoded").
		BodyString(fmt.Sprintf("_Password=%s&_QString=&_UserName=%s&__RequestVerificationToken=.*", url.QueryEscape(validPassword), validUsername)).
		Reply(http.StatusOK).
		AddHeader("Set-Cookie", fmt.Sprintf("ASP.NET_SessionId=%s; path=/; HttpOnly", fakeSessionId)).
		AddHeader("Set-Cookie", fmt.Sprintf("__RequestVerificationToken=%s; path=/; HttpOnly", fakeRequestVerificationToken)).
		AddHeader("Set-Cookie", fmt.Sprintf(".ASPXAUTH=%s; path=/; HttpOnly", fakeAuthCookie))
}

func gockRegisterHomePageLoggedIn() error {
	mockHome, err := os.Open("testdata/home_page_logged_in.html")
	if err != nil {
		return errors.New("Failed to open mock home page: " + err.Error())
	}

	gock.New("https://s.amizone.net").
		Get("/Home").
		MatchHeader("User-Agent", ".*").
		MatchHeader("Referer", "https://s.amizone.net").
		MatchHeader("Cookie", fmt.Sprintf("ASP.NET_SessionId=%s", fakeSessionId)).
		MatchHeader("Cookie", fmt.Sprintf(".ASPXAUTH=%s", fakeAuthCookie)).
		MatchHeader("Cookie", fmt.Sprintf("__RequestVerificationToken=%s", fakeRequestVerificationToken)).
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockHome)
	return nil
}
