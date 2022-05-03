package amizone

import (
	"amizone/amizone/internal"
	"amizone/amizone/internal/models"
	"amizone/amizone/internal/parse"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	BaseUrl = "https://" + internal.AmizoneDomain

	loginRequestEndpoint     = "/"
	attendancePageEndpoint   = "/Home"
	scheduleEndpointTemplate = "/Calendar/home/GetDiaryEvents?start=%s&end=%s"

	scheduleEndpointTimeFormat = "2006-01-02"
	scheduleJsonTimeFormat     = "2006/01/02 03:04:05 PM"

	verificationTokenName = "__RequestVerificationToken"

	ErrBadClient              = "the http client passed must have a cookie jar, or be nil"
	ErrFailedToVisitPage      = "failed to visit page"
	ErrFailedToReadResponse   = "failed to read response body"
	ErrFailedLogin            = "failed to login"
	ErrInvalidCredentials     = ErrFailedLogin + ": invalid credentials"
	ErrInternalFailure        = "internal failure"
	ErrFailedToComposeRequest = ErrInternalFailure + ": failed to compose request"
	ErrFailedToParsePage      = ErrInternalFailure + ": failed to parse page"
)

type Credentials struct {
	Username string
	Password string
}

// amizoneClient is the main struct for the amizone package, exposing
// the entire API surface for the website as implemented here.
type amizoneClient struct {
	client      *http.Client
	credentials *Credentials
	muLogin     struct {
		sync.Mutex
		lastAttempt time.Time
		didLogin    bool
	}
}

// DidLogin returns true if the client ever successfully logged in.
func (a *amizoneClient) DidLogin() bool {
	a.muLogin.Lock()
	defer a.muLogin.Unlock()
	return a.muLogin.didLogin
}

// Interface compliance constraint for amizoneClient
var _ ClientInterface = &amizoneClient{}

// NewClient create a new amizoneClient instance with Credentials passed, then attempts to log in to the website.
// The *http.Client parameter can be nil, in which case a default client will be created in its place.
// To get a non-logged in client, pass empty credentials, ala Credentials{}.
func NewClient(cred Credentials, httpClient *http.Client) (*amizoneClient, error) {
	if httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			klog.Error("failed to create cookiejar for the amizone client. this is a bug, please report it.")
			return nil, errors.New(ErrInternalFailure)
		}
		httpClient = &http.Client{Jar: jar}
	}

	if jar := httpClient.Jar; jar == nil {
		klog.Error("amizone.NewClient called with a jar-less http client")
		return nil, errors.New(ErrBadClient)
	}

	client := &amizoneClient{
		client:      httpClient,
		credentials: &cred,
	}

	// We don't try to log in if empty credentials were passed
	if cred == (Credentials{}) {
		return client, nil
	}

	return client, client.login()
}

// login attempts to log in to Amizone with the credentials passed to the amizoneClient and a scrapped
// "__RequestVerificationToken" value.
func (a *amizoneClient) login() error {
	a.muLogin.Lock()
	defer a.muLogin.Unlock()

	if time.Now().Sub(a.muLogin.lastAttempt) < time.Minute*2 {
		return nil
	}

	// Our last attempt is NOW
	a.muLogin.lastAttempt = time.Now()

	// Amizone uses a "verification" token for logins -- we try to retrieve this from the login form page
	verToken := func() string {
		response, err := a.doRequest(false, http.MethodGet, "/", nil)
		if err != nil {
			klog.Errorf("login: %s", err.Error())
			return ""
		}
		return parse.VerificationToken(response.Body)
	}()

	if verToken == "" {
		klog.Error("login: failed to retrieve verification token from the login page")
		return errors.New(fmt.Sprintf("%s: %s", ErrFailedLogin, ErrFailedToParsePage))
	}

	loginRequestData := func() (v url.Values) {
		v = url.Values{}
		v.Set(verificationTokenName, verToken)
		v.Set("_UserName", a.credentials.Username)
		v.Set("_Password", a.credentials.Password)
		v.Set("_QString", "")
		return
	}()

	loginResponse, err := a.doRequest(false, http.MethodPost, loginRequestEndpoint, strings.NewReader(loginRequestData.Encode()))
	if err != nil {
		klog.Warningf("Something went wrong while making the login request: ", err.Error())
		return errors.New(fmt.Sprintf("%s: %s", ErrFailedLogin, err.Error()))
	}

	// The login request should redirect our request to the home page with a 302 "found" status code.
	// If we're instead redirected to the login page, we've failed to log in because of invalid credentials
	if loginResponse.Request.URL.Path == loginRequestEndpoint {
		return errors.New(ErrInvalidCredentials)
	}

	if loggedIn := parse.LoggedIn(loginResponse.Body); !loggedIn {
		klog.Error("Login failed. Possible reasons: something broke.")
		return errors.New(ErrFailedLogin)
	}

	// We need to check if the right tokens are here in the cookie jar to make sure we're logged in
	if !internal.IsLoggedIn(a.client) {
		klog.Error("Login failed. Possible reasons: something broke.")
		return errors.New(ErrFailedLogin)
	}

	a.muLogin.didLogin = true

	return nil
}

// GetAttendance retrieves, parses and returns attendance data from Amizone
func (a *amizoneClient) GetAttendance() (models.AttendanceRecord, error) {
	response, err := a.doRequest(true, http.MethodGet, attendancePageEndpoint, nil)
	if err != nil {
		klog.Warningf("request (attendance): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	attendanceRecord, err := parse.Attendance(response.Body)
	if err != nil {
		klog.Errorf("parse (attendance): %s", err.Error())
		return nil, errors.New(ErrFailedToParsePage)
	}

	return attendanceRecord, nil
}

func (a *amizoneClient) GetClassSchedule(date Date) (models.ClassSchedule, error) {
	timeFrom := time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, time.UTC)
	timeTo := timeFrom.Add(time.Hour * 24)

	var schedule models.ClassSchedule
	var unmarshalErr error

	// amizoneEntry is the JSON format we expect from the Amizone
	type amizoneEntry struct {
		Type       string `json:"sType"` // "C" for course, "E" for event, "H" for holiday
		CourseName string `json:"title"`
		CourseCode string `json:"CourseCode"`
		Faculty    string `json:"FacultyName"`
		Room       string `json:"RoomNo"`
		Start      string `json:"start"` // Start and end keys are in the format "YYYY-MM-DD HH:MM:SS"
		End        string `json:"end"`
	}
	var amizoneSchedule []amizoneEntry

	c := internal.GetNewColly(a.client, true)
	c.OnResponse(func(r *colly.Response) {
		unmarshalErr = json.Unmarshal(r.Body, &amizoneSchedule)
		if unmarshalErr != nil {
			klog.Errorf("Failed to unmarshall JSON response from Amizone: %s. Are we logged in?", unmarshalErr.Error())
			return
		}

		for _, entry := range amizoneSchedule {
			// Only add entries that are of type "C" (class)
			if entry.Type != "C" {
				continue
			}

			timeParserFunc := func(timeStr string) time.Time {
				t, err := time.Parse(scheduleJsonTimeFormat, timeStr)
				if err != nil {
					klog.Warning("Failed to parse time for course %s: %s", entry.CourseCode, err.Error())
					return time.Unix(0, 0)
				}
				return t
			}

			class := &models.ScheduledClass{
				Course: &models.Course{
					Code: entry.CourseCode,
					Name: entry.CourseName,
				},
				StartTime: timeParserFunc(entry.Start),
				EndTime:   timeParserFunc(entry.End),
				Faculty:   entry.Faculty,
				Room:      entry.Room,
			}

			schedule = append(schedule, class)
		}
	})

	err := c.Visit(BaseUrl + fmt.Sprintf(scheduleEndpointTemplate, timeFrom.Format(scheduleEndpointTimeFormat), timeTo.Format(scheduleEndpointTimeFormat)))
	if err != nil {
		klog.Error("Something went wrong while visiting the schedule endpoint: " + err.Error())
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToVisitPage, err.Error()))
	}

	// We sort the parsed schedule by start time -- because the Amizone events endpoint does not guarantee order.
	schedule.Sort()

	return schedule, nil
}
