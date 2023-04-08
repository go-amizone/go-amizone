package amizone_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ditsuke/go-amizone/amizone"
	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/models"
	. "github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

// === Test setup helpers ===

type Empty struct{}

// DummyMatcher is a matcher for the Empty datatype that does exactly nothing,
// for when the function to be tested returns nothing.
func DummyMatcher[T any](_ T, _ *WithT) {
}

// DummySetup is used when a test requires no setup.
func DummySetup(_ *WithT) {
}

func NoError(err error, g *WithT) {
	g.Expect(err).ToNot(HaveOccurred())
}

// / TestCase is a generic type to reduce test boilerplate
type TestCase[D any, I any] struct {
	name        string
	client      *amizone.Client
	setup       func(g *WithT)
	input       I
	dataMatcher func(data D, g *WithT)
	errMatcher  func(err error, g *WithT)
}

// Sanity check testcase, since the go type system won't do it for us 😭
func (c *TestCase[D, I]) sanityCheck(g *WithT) {
	g.Expect(c.setup).ToNot(BeNil(), "setup function must not be nil")
	g.Expect(c.dataMatcher).ToNot(BeNil(), "data matcher function must not be nil")
	g.Expect(c.errMatcher).ToNot(BeNil(), "error matcher function must not be nil")
}

// === Test helpers ===

// toJSON converts a struct to a JSON string.
func toJSON[T any](t T, g *WithT) string {
	s, err := json.Marshal(t)
	g.Expect(err).ToNot(HaveOccurred(), "marshall json")
	return string(s)
}

// @todo: implement test cases to test behavior when:
// - Amizone is not reachable
// - Amizone is reachable but login fails (invalid credentials, etc?)
func TestNewClient(t *testing.T) {
	g := NewGomegaWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	err := mock.GockRegisterLoginPage()
	g.Expect(err).ToNot(HaveOccurred(), "failed to register login page mock")
	err = mock.GockRegisterLoginRequest()
	g.Expect(err).ToNot(HaveOccurred(), "failed to register login request mock")

	jar, err := cookiejar.New(nil)
	g.Expect(err).ToNot(HaveOccurred(), "failed to create cookie jar")

	httpClient := &http.Client{Jar: jar}
	gock.InterceptClient(httpClient)

	c := amizone.Credentials{
		Username: mock.ValidUser,
		Password: mock.ValidPass,
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

	setupNetworking()
	t.Cleanup(teardown)

	nonLoggedInClient := getNonLoggedInClient(g)
	loggedInClient := getLoggedInClient(g)

	gock.Clean()

	testCases := []struct {
		name              string
		amizoneClient     *amizone.Client
		setup             func(g *WithT)
		attendanceMatcher func(g *WithT, attendance models.AttendanceRecords)
		errorMatcher      func(g *WithT, err error)
	}{
		{
			name:          "Logged in, expecting retrieval",
			amizoneClient: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterHomePageLoggedIn()
				g.Expect(err).ToNot(HaveOccurred())
			},
			attendanceMatcher: func(g *WithT, attendance models.AttendanceRecords) {
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
				err := mock.GockRegisterUnauthenticatedGet("/Home")
				g.Expect(err).ToNot(HaveOccurred())
			},
			attendanceMatcher: func(g *WithT, attendance models.AttendanceRecords) {
				g.Expect(attendance).To(BeEmpty())
			},
			errorMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedLogin))
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			t.Cleanup(setupNetworking)

			c.setup(g)

			attendance, err := c.amizoneClient.GetAttendance()
			c.attendanceMatcher(g, attendance)
			c.errorMatcher(g, err)
		})
	}
}

func TestClient_GetSemesters(t *testing.T) {
	g := NewGomegaWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	testCases := []struct {
		name             string
		client           *amizone.Client
		setup            func(g *WithT)
		semestersMatcher func(g *WithT, semesters models.SemesterList)
		errMatcher       func(g *WithT, err error)
	}{
		{
			name:   "client is logged in and amizone returns a (mock) courses page",
			client: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterCurrentCoursesPage()
				g.Expect(err).ToNot(HaveOccurred())
			},
			semestersMatcher: func(g *WithT, semesters models.SemesterList) {
				g.Expect(semesters).To(HaveLen(4))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:   "client is not logged in and amizone returns the login page",
			client: nonLoggedInClient,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterLoginPage()).ToNot(HaveOccurred())
			},
			semestersMatcher: func(g *WithT, semesters models.SemesterList) {
				g.Expect(semesters).To(HaveLen(0))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCase.setup(g)

			semesters, err := testCase.client.GetSemesters()
			testCase.errMatcher(g, err)
			testCase.semestersMatcher(g, semesters)
		})
	}
}

func TestClient_GetCourses(t *testing.T) {
	g := NewWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	testCases := []struct {
		name           string
		client         *amizone.Client
		semesterRef    string
		setup          func(g *WithT)
		coursesMatcher func(g *WithT, courses models.Courses)
		errMatcher     func(g *WithT, err error)
	}{
		{
			name:        "amizone client is logged in, we ask for semester 1, return mock courses page on expected POST",
			client:      loggedInClient,
			semesterRef: "1",
			setup: func(g *WithT) {
				err := mock.GockRegisterSemesterCoursesRequest("1")
				g.Expect(err).ToNot(HaveOccurred())
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(8))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:        "amizone client is logged in, we ask for semester 2, return mock courses page on expected POST",
			client:      loggedInClient,
			semesterRef: "2",
			setup: func(g *WithT) {
				err := mock.GockRegisterSemesterCoursesRequest("2")
				g.Expect(err).ToNot(HaveOccurred())
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(8))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:        "amizone client is not logged in, returns login page on request",
			client:      nonLoggedInClient,
			semesterRef: "3",
			setup: func(g *WithT) {
				//err := mock.GockRegisterLoginPage()
				//g.Expect(err).ToNot(HaveOccurred())
				err := mock.GockRegisterUnauthenticatedGet("/")
				g.Expect(err).ToNot(HaveOccurred())
				mock.GockRegisterUnauthenticatedPost("/CourseListSemWise", url.Values{"sem": []string{"3"}}.Encode(), strings.NewReader("<no></no>"))
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(0))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).ToNot(ContainSubstring(amizone.ErrFailedToVisitPage))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCase.setup(g)

			courses, err := testCase.client.GetCourses(testCase.semesterRef)
			testCase.errMatcher(g, err)
			testCase.coursesMatcher(g, courses)
		})
	}
}

func TestClient_GetCurrentCourses(t *testing.T) {
	g := NewWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	testCases := []struct {
		name           string
		client         *amizone.Client
		setup          func(g *WithT)
		coursesMatcher func(g *WithT, courses models.Courses)
		errMatcher     func(g *WithT, err error)
	}{
		{
			name:   "amizone client is logged in and returns the (mock) courses page",
			client: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterCurrentCoursesPage()
				g.Expect(err).ToNot(HaveOccurred())
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(8))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:   "amizone client is logged is and returns the (mock) sem-wise courses page",
			client: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterSemWiseCoursesPage()
				g.Expect(err).ToNot(HaveOccurred())
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(8))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:   "amizone client is not logged in and returns the login page",
			client: nonLoggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterUnauthenticatedGet("/")
				g.Expect(err).ToNot(HaveOccurred())
			},
			coursesMatcher: func(g *WithT, courses models.Courses) {
				g.Expect(courses).To(HaveLen(0))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCase.setup(g)

			courses, err := testCase.client.GetCurrentCourses()
			testCase.errMatcher(g, err)
			testCase.coursesMatcher(g, courses)
		})
	}
}

func TestClient_GetProfile(t *testing.T) {
	g := NewWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	testCases := []struct {
		name           string
		client         *amizone.Client
		setup          func(g *WithT)
		profileMatcher func(g *WithT, profile *models.Profile)
		errMatcher     func(g *WithT, err error)
	}{
		{
			name:   "amizone client logged in and returns the (mock) profile page",
			client: loggedInClient,
			setup: func(g *WithT) {
				err := mock.GockRegisterProfilePage()
				g.Expect(err).ToNot(HaveOccurred())
			},
			profileMatcher: func(g *WithT, profile *models.Profile) {
				g.Expect(profile).To(Equal(&models.Profile{
					Name:               mock.StudentName,
					EnrollmentNumber:   mock.StudentEnrollmentNumber,
					EnrollmentValidity: mock.StudentIDValidity.Time(),
					DateOfBirth:        mock.StudentDOB.Time(),
					Batch:              mock.StudentBatch,
					Program:            mock.StudentProgram,
					BloodGroup:         mock.StudentBloodGroup,
					IDCardNumber:       mock.StudentIDCardNumber,
					UUID:               mock.StudentUUID,
				}))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
		{
			name:   "amizone client is not logged in and returns the login page",
			client: nonLoggedInClient,
			setup: func(g *WithT) {
				_ = mock.GockRegisterUnauthenticatedGet("/IDCard")
			},
			profileMatcher: func(g *WithT, profile *models.Profile) {
				g.Expect(profile).To(BeNil())
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedLogin))
			},
		},
	}

	for _, testCases := range testCases {
		t.Run(testCases.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCases.setup(g)

			profile, err := testCases.client.GetProfile()
			testCases.errMatcher(g, err)
			testCases.profileMatcher(g, profile)
		})
	}
}
func macStringtoMac(a string, g *WithT) net.HardwareAddr {
	addr, err := net.ParseMAC(a)
	g.Expect(err).ToNot(HaveOccurred())
	return addr
}

func TestClient_GetWifiMacInfo(t *testing.T) {
	g := NewWithT(t)

	setupNetworking()
	t.Cleanup(teardown)

	loggedInClient := getLoggedInClient(g)
	_ = getNonLoggedInClient(g)

	testCases := []struct {
		name        string
		client      *amizone.Client
		setup       func(g *WithT)
		infoMatcher func(g *WithT, info *models.WifiMacInfo)
		errMatcher  func(g *WithT, err error)
	}{
		{
			name:   "amizone returns macs as usual",
			client: loggedInClient,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
			},
			infoMatcher: func(g *WithT, info *models.WifiMacInfo) {
				g.Expect(info).ToNot(BeNil())
				g.Expect(info.RegisteredAddresses).To(HaveLen(2))
				g.Expect(toJSON(info, g)).To(MatchJSON(`{"RegisteredAddresses":["VQQt576k","/dUUGAyL"],"Slots":2,"FreeSlots":0}`))
			},
			errMatcher: func(g *WithT, err error) {
				g.Expect(err).ToNot(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCase.setup(g)

			info, err := testCase.client.GetWifiMacInfo()
			testCase.errMatcher(g, err)
			testCase.infoMatcher(g, info)
		})
	}
}

func TestClient_RegisterWifiMac(t *testing.T) {
	setupNetworking()
	t.Cleanup(teardown)
	g := NewWithT(t)

	type RegisterMacArgs = struct {
		A net.HardwareAddr
		O bool
	}

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	macNew := macStringtoMac(mock.ValidMacNew, g)

	infoOneShot, err := mock.WifiPageOneSlot.Open()
	g.Expect(err).ToNot(HaveOccurred())
	verificationToken := parse.VerificationToken(infoOneShot)

	testCases := []TestCase[Empty, RegisterMacArgs]{
		{
			// Go's net.HardwareAddr is not guaranteed to be valid :smiles_in_pain:
			name:   "client: logged in; mac: invalid; bypass: false",
			client: loggedInClient,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterHomePageLoggedIn()).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
			},
			input:       RegisterMacArgs{A: net.HardwareAddr{}, O: false},
			dataMatcher: DummyMatcher[Empty],
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrInvalidMac))
			},
		},
		{
			name:        "client: logged in; mac: valid; free_slots: none; bypass: false",
			client:      loggedInClient,
			dataMatcher: DummyMatcher[Empty],
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrNoMacSlots))
			},
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
			},
			input: RegisterMacArgs{A: macNew, O: false},
		},
		{
			name:        "client: logged in; mac: valid; free_slots: none; bypass: true",
			client:      loggedInClient,
			dataMatcher: DummyMatcher[Empty],
			errMatcher:  NoError,
			input:       RegisterMacArgs{A: macNew, O: true},
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterWifiRegistration(url.Values{
					"__RequestVerificationToken": {verificationToken},
					"Amizone_Id":                 {mock.ValidUser},
					"Mac1":                       {mock.ValidMac1},
					"Mac2":                       {mock.ValidMacNew},
					"Name":                       {"DoesntMatter"},
				}))
			},
		},
		{
			name:        "client: logged in; mac: valid; free slots: 1, bypass: false",
			client:      loggedInClient,
			input:       RegisterMacArgs{A: macNew, O: false},
			dataMatcher: DummyMatcher[Empty],
			errMatcher:  NoError,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfoOneSlot()).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterWifiRegistration(url.Values{
					"__RequestVerificationToken": {verificationToken},
					"Amizone_Id":                 {mock.ValidUser},
					"Mac1":                       {mock.ValidMac1},
					"Mac2":                       {mock.ValidMacNew},
					"Name":                       {"DoesntMatter"},
				}))
			},
		},
		{
			name:        "client: logged in; mac: valid; free_slots: 1; bypass: true",
			client:      loggedInClient,
			input:       RegisterMacArgs{A: macNew, O: true},
			dataMatcher: DummyMatcher[Empty],
			errMatcher:  NoError,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfoOneSlot()).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterWifiRegistration(url.Values{
					"__RequestVerificationToken": {verificationToken},
					"Amizone_Id":                 {mock.ValidUser},
					"Mac1":                       {mock.ValidMac1},
					"Mac2":                       {mock.ValidMacNew},
					"Name":                       {"DoesntMatter"},
				}))
			},
		},
		{
			name:        "client is logged in, mac already exists",
			client:      loggedInClient,
			input:       RegisterMacArgs{A: macStringtoMac(mock.ValidMac2, g), O: false},
			dataMatcher: DummyMatcher[Empty],
			errMatcher:  NoError,
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
				// We don't expect a registration request
			},
		},
		{
			name:   "client not logged in, returns error",
			client: nonLoggedInClient,
			input:  RegisterMacArgs{A: macStringtoMac(mock.ValidMac1, g), O: false},
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterWifiInfo()).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterUnauthenticatedGet("/Home")).ToNot(HaveOccurred())
				g.Expect(mock.GockRegisterUnauthenticatedGet("RegisterForWifi/mac/MacRegistration")).ToNot(HaveOccurred())
			},
			dataMatcher: DummyMatcher[Empty],
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedLogin))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			g := NewWithT(t)
			t.Cleanup(setupNetworking)
			testCase.sanityCheck(g)
			testCase.setup(g)
			err := testCase.client.RegisterWifiMac(testCase.input.A, testCase.input.O)
			testCase.errMatcher(err, g)
		})
	}
}

func TestClient_RemoveWifiMac(t *testing.T) {
	// TODO
	setupNetworking()
	t.Cleanup(teardown)
	g := NewWithT(t)

	type RemoveWifiArgs = struct {
		A net.HardwareAddr
	}

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	testCases := []TestCase[Empty, RemoveWifiArgs]{
		{
			name:   "mac address is invalid",
			client: loggedInClient,
			setup:  DummySetup,
			input:  RemoveWifiArgs{A: net.HardwareAddr{}},
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrInvalidMac))
			},
			dataMatcher: DummyMatcher[Empty],
		},
		{
			name:   "amizone is unreachable",
			client: loggedInClient,
			setup:  DummySetup,
			input:  RemoveWifiArgs{A: macStringtoMac(mock.ValidMac1, g)},
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedToVisitPage))
			},
			dataMatcher: DummyMatcher[Empty],
		},
		{
			name:   "client is not logged in",
			client: nonLoggedInClient,
			setup:  DummySetup,
			input:  RemoveWifiArgs{A: macStringtoMac(mock.ValidMac1, g)},
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedLogin))
			},
			dataMatcher: DummyMatcher[Empty],
		},
		{
			name:   "parser breaks when amizone changes something",
			client: loggedInClient,
			setup: func(g *WithT) {
				// Return some random other page
				g.Expect(
					mock.GockRegisterWifiMacDeletion(
						map[string]string{
							"username":   mock.ValidUser,
							"Amizone_Id": mock.ValidMac2,
						},
						// Send some unexpected page back
						mock.CoursesPage,
					),
				)
			},
			input: RemoveWifiArgs{A: macStringtoMac(mock.ValidMac2, g)},
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedToParsePage))
			},
			dataMatcher: DummyMatcher[Empty],
		},
		{
			name:   "everything goes ok",
			client: loggedInClient,
			setup: func(g *WithT) {
				// Return some random other page
				g.Expect(
					mock.GockRegisterWifiMacDeletion(
						map[string]string{
							"username":   mock.ValidUser,
							"Amizone_Id": mock.ValidMac2,
						},
						mock.WifiPageOneSlot,
					),
				)
			},
			input:       RemoveWifiArgs{A: macStringtoMac(mock.ValidMac2, g)},
			errMatcher:  NoError,
			dataMatcher: DummyMatcher[Empty],
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(setupNetworking)
			g := NewWithT(t)

			testCase.sanityCheck(g)
			testCase.setup(g)
			err := testCase.client.RemoveWifiMac(testCase.input.A)
			testCase.errMatcher(err, g)
		})
	}
}

// Test out amizone.GetClassSchedule in the style of the other tests written
func TestClient_GetClassSchedule(t *testing.T) {
	setupNetworking()
	t.Cleanup(teardown)
	g := NewWithT(t)

	type GetClassScheduleArgs = struct {
		year  int
		month time.Month
		day   int
	}

	loggedInClient := getLoggedInClient(g)
	nonLoggedInClient := getNonLoggedInClient(g)

	standardDate := GetClassScheduleArgs{year: 2023, month: time.April, day: 1}
	standardDatePlusOne := GetClassScheduleArgs{year: 2023, month: time.April, day: 2}
	fmtDate := func(args GetClassScheduleArgs) string {
		return fmt.Sprintf("%02d-%02d-%02d", args.year, args.month, args.day)
	}

	testCases := []TestCase[models.ClassSchedule, GetClassScheduleArgs]{
		{
			name:   "client is not logged in",
			client: nonLoggedInClient,
			input:  standardDate,
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedLogin))
			},
			dataMatcher: DummyMatcher[models.ClassSchedule],
			setup:       DummySetup,
		},
		{
			name:   "amizone doesn't send back any events",
			client: loggedInClient,
			input:  standardDate,
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).ToNot(HaveOccurred())
			},
			dataMatcher: func(data models.ClassSchedule, g *WithT) {
				g.Expect(data).To(BeEmpty())
			},
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterCalendarEndpoint(fmtDate(standardDate), fmtDate(standardDatePlusOne), mock.DiaryEventsNone)).ToNot(HaveOccurred())
			},
		},
		{
			name:   "amizone's response cannot be parsed (no longer json)",
			client: loggedInClient,
			input:  standardDate,
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(amizone.ErrFailedToParsePage))
			},
			dataMatcher: DummyMatcher[models.ClassSchedule],
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterCalendarEndpoint(fmtDate(standardDate), fmtDate(standardDatePlusOne), mock.CoursesPage)).ToNot(HaveOccurred())
			},
		},
		{
			name:   "amizone sends back response with events",
			client: loggedInClient,
			input:  standardDate,
			errMatcher: func(err error, g *WithT) {
				g.Expect(err).ToNot(HaveOccurred())
			},
			dataMatcher: func(schedule models.ClassSchedule, g *WithT) {
				g.Expect(schedule).To(HaveLen(3))
				sb := strings.Builder{}
				_ = json.NewEncoder(&sb).Encode(schedule)
				g.Expect(sb.String()).To(MatchJSON(`[{"Course":{"Code":"IT414","Name":"SS"},"StartTime":"2023-04-01T12:15:00Z","EndTime":"2023-04-01T13:10:00Z","Faculty":"DRS[2434]","Room":"E1-309","Attended":2},{"Course":{"Code":"IT301","Name":"SE"},"StartTime":"2023-04-01T12:15:00Z","EndTime":"2023-04-01T13:10:00Z","Faculty":"DRG[2397],DSKD[2436]","Room":"E1-000","Attended":1},{"Course":{"Code":"CSE304","Name":"CC"},"StartTime":"2023-04-01T13:15:00Z","EndTime":"2023-04-01T14:10:00Z","Faculty":"DAG[307870]","Room":"E1-000","Attended":0}]`))
				g.Expect(schedule[0].Attended).To(Equal(models.AttendanceStateAbsent))
				g.Expect(schedule[1].Attended).To(Equal(models.AttendanceStatePresent))
				g.Expect(schedule[2].Attended).To(Equal(models.AttendanceStatePending))
			},
			setup: func(g *WithT) {
				g.Expect(mock.GockRegisterCalendarEndpoint(fmtDate(standardDate), fmtDate(standardDatePlusOne), mock.DiaryEventsSmallJSON)).ToNot(HaveOccurred())
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Cleanup(setupNetworking)
			g := NewWithT(t)

			testCase.sanityCheck(g)
			testCase.setup(g)
			classes, err := testCase.client.ClassSchedule(testCase.input.year, testCase.input.month, testCase.input.day)
			testCase.errMatcher(err, g)
			testCase.dataMatcher(classes, g)
		})
	}
}

// Test utilities

// setupNetworking tears down any existing network mocks and sets up gock anew to intercept network
// calls and disable real network calls.
func setupNetworking() {
	// tear everything all routes down
	teardown()
	gock.Intercept()
	gock.DisableNetworking()
}

// teardown disables all networking restrictions and mock routes registered with gock for unit testing.
func teardown() {
	gock.Clean()
	gock.Off()
	gock.EnableNetworking()
}

func getNonLoggedInClient(g *GomegaWithT) *amizone.Client {
	client, err := amizone.NewClient(amizone.Credentials{}, nil)
	g.Expect(err).ToNot(HaveOccurred())
	return client
}

func getLoggedInClient(g *GomegaWithT) *amizone.Client {
	err := mock.GockRegisterLoginPage()
	g.Expect(err).ToNot(HaveOccurred(), "failed to register mock login page")
	err = mock.GockRegisterLoginRequest()
	g.Expect(err).ToNot(HaveOccurred(), "failed to register mock login request")

	client, err := amizone.NewClient(amizone.Credentials{
		Username: mock.ValidUser,
		Password: mock.ValidPass,
	}, nil)
	g.Expect(err).ToNot(HaveOccurred(), "failed to setup mock logged-in client")
	return client
}
