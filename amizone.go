package amizone

import (
	"amizone/internal"
	"amizone/internal/models"
	"amizone/internal/parse"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"io"
	"io/ioutil"
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

	ErrFailedToVisitPage     = "failed to visit page"
	ErrFailedToReadResponse  = "failed to read response body"
	ErrFailedToParsePage     = "failed to parse page"
	ErrFailedToRetrieveToken = "failed to retrieve verification token"
	ErrFailedLogin           = "failed to login"

	errFailedToComposeRequest = "failed to compose request"
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
	didLogin    bool
	muLogin     struct {
		sync.Mutex
		lastAttempt time.Time
	}
}

// DidLogin returns true if the client ever successfully logged in.
func (a *amizoneClient) DidLogin() bool {
	return a.didLogin
}

// doRequest is an internal http request helper to simplify making requests.
// This method takes care of both composing requests, setting custom headers and such as needed.
// If tryLogin is true, the client will attempt to log in if it is not already logged in.
// method must be a valid http request method.
// endpoint must be relative to BaseUrl.
func (a *amizoneClient) doRequest(tryLogin bool, method string, endpoint string, body io.Reader) (*http.Response, error) {
	// Login now if we didn't log in at instantiation.
	if tryLogin && !a.didLogin && *a.credentials != (Credentials{}) {
		if err := a.login(); err != nil {
			return nil, errors.New(ErrFailedLogin)
		}
		tryLogin = false // We don't want to attempt another login.
	}

	req, err := http.NewRequest(method, BaseUrl+endpoint, body)
	if err != nil {
		klog.Errorf("%s: %s", errFailedToComposeRequest, err)
		return nil, errors.New(errFailedToComposeRequest)
	}

	req.Header.Set("User-Agent", internal.Firefox99UserAgent)
	// Amizone uses the referrer to authenticate requests on top of the actual AUTH/session cookies.
	req.Header.Set("Referer", BaseUrl+"/")
	if method == http.MethodPost { // We assume a POST request means submitting a form.
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	response, err := a.client.Do(req)
	if err != nil {
		klog.Errorf(fmt.Sprintf("%s: %s", ErrFailedToVisitPage, err))
		return nil, errors.New(ErrFailedToVisitPage)
	}

	// Read the response into a byte array, so we can reuse it.
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return response, errors.New(ErrFailedToReadResponse)
	}
	_ = response.Body.Close()

	response.Body = ioutil.NopCloser(bytes.NewReader(responseBody))

	// If we're directed to try log-ins and the parser determines we're not logged in, we retry.
	if tryLogin && *a.credentials != (Credentials{}) && !parse.LoggedIn(bytes.NewReader(responseBody)) {
		if err := a.login(); err != nil {
			return nil, errors.New(ErrFailedLogin)
		}
		return a.doRequest(false, method, endpoint, body)
	}

	return response, nil
}

// Interface compliance constraint for amizoneClient
var _ ClientInterface = &amizoneClient{}

// NewClient create a new amizoneClient instance with Credentials passed, then attempts to log in to the website.
// The *http.Client parameter can be nil, in which case a default client will be created in its place.
// To get a non-logged in client, pass empty credentials, ala Credentials{}.
func NewClient(creds Credentials, httpClient *http.Client) (*amizoneClient, error) {
	if httpClient == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, errors.New("failed to create cookie jar for httpClient: " + err.Error())
		}
		httpClient = &http.Client{Jar: jar}
	}

	if jar := httpClient.Jar; jar == nil {
		klog.Error("Credentials.NewClient called with a Jarless client.")
		return nil, errors.New("must pass a http.Client with a cookie jar or pass a nil client")
	}

	client := &amizoneClient{
		client:      httpClient,
		credentials: &creds,
	}

	if creds == (Credentials{}) {
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
			klog.Errorf("login: %s", err)
			return ""
		}
		return parse.VerificationToken(response.Body)
	}()

	if verToken == "" {
		klog.Error("Failed to retrieve verification token from login page. What's up?")
		return errors.New(ErrFailedToRetrieveToken)
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
		klog.Error("Something went wrong while posting login data: ", err.Error())
		return errors.New(fmt.Sprintf("%s: %s", ErrFailedLogin, err.Error()))
	}

	if loggedIn := parse.LoggedIn(loginResponse.Body); !loggedIn {
		klog.Error("Login failed. Are these credentials valid?")
		return errors.New(ErrFailedLogin)
	}

	// We need to check if the right tokens are here in the cookie jar to make sure we're logged in
	if !internal.IsLoggedIn(a.client) {
		klog.Error("Login failed. Are these credentials valid?")
		return errors.New(ErrFailedLogin)
	}

	a.didLogin = true

	return nil
}

// GetAttendance retrieves, parses and returns attendance data from Amizone
func (a *amizoneClient) GetAttendance() (models.AttendanceRecord, error) {
	response, err := a.doRequest(true, http.MethodGet, attendancePageEndpoint, nil)
	if err != nil {
		klog.Errorf("get_attendance: %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	attendanceRecord, err := parse.Attendance(response.Body)
	if err != nil {
		klog.Errorf("%s (attendance): %s", ErrFailedToParsePage, err.Error())
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
