package amizone_test

import (
	"GoFriday/lib/amizone"
	"GoFriday/lib/amizone/internal/mock"
	"GoFriday/lib/amizone/internal/models"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"net/http/cookiejar"
	"testing"
)

type amizoneClientInterface interface {
	GetAttendance() (models.AttendanceRecord, error)
}

// @todo: implement test cases to test behavior when:
// - Amizone is not reachable
// - Amizone is reachable but login fails (invalid credentials, etc?)
func TestNewClient(t *testing.T) {
	g := NewGomegaWithT(t)

	gock.DisableNetworking()
	defer gock.Off()
	defer gock.EnableNetworking()

	err := mock.GockRegisterLoginPage()
	g.Expect(err).ToNot(HaveOccurred(), "failed to register login page mock")
	err = mock.GockRegisterLoginRequest(mock.AmizoneUsername, mock.AmizonePassword)
	g.Expect(err).ToNot(HaveOccurred(), "failed to register login request mock")

	jar, err := cookiejar.New(nil)
	g.Expect(err).ToNot(HaveOccurred(), "Failed to create cookie jar")

	httpClient := &http.Client{Jar: jar}
	gock.InterceptClient(httpClient)

	c := amizone.Credentials{
		Username: mock.AmizoneUsername,
		Password: mock.AmizonePassword,
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
		err := mock.GockRegisterLoginPage()
		g.Expect(err).ToNot(HaveOccurred(), "Failed to register mock login page")
		err = mock.GockRegisterLoginRequest(mock.AmizoneUsername, mock.AmizonePassword)
		g.Expect(err).ToNot(HaveOccurred(), "Failed to register mock login request")

		loggedInClient, err = amizone.NewClient(amizone.Credentials{
			Username: mock.AmizoneUsername,
			Password: mock.AmizonePassword,
		}, nil)
		g.Expect(err).ToNot(HaveOccurred(), "set up mocked logged-in amizone client")
	}()

	testCases := []struct {
		name              string
		amizoneClient     amizoneClientInterface
		setup             func(g *WithT)
		attendanceMatcher func(g *WithT, attendance models.AttendanceRecord)
		errorMatcher      func(g *WithT, err error)
	}{
		{
			name:          "Logged in, expecting retrieval",
			amizoneClient: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterHomePageLoggedIn()
				g.Expect(err).ToNot(HaveOccurred())
			},
			attendanceMatcher: func(g *WithT, attendance models.AttendanceRecord) {
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
				err := mock.GockRegisterHomePageLoggedIn()
				g.Expect(err).ToNot(HaveOccurred())
				gock.New("https://s.amizone.net").
					Get("/Home").
					Reply(http.StatusOK).
					BodyString("<html><body>Forbidden -- No Records for you</body></html>")
			},
			attendanceMatcher: func(g *WithT, attendance models.AttendanceRecord) {
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
