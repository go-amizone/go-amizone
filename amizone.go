package amizone

import (
	"amizone/internal"
	"amizone/internal/models"
	"amizone/internal/parse"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"io"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
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

	ErrFailedAttendanceRetrieval = "failed to retrieve attendance"
	ErrFailedToVisitPage         = "failed to visit page"
	ErrFailedToParsePage         = "failed to parse page"
	ErrFailedToRetrieveToken     = "failed to retrieve verification token"
	ErrFailedLogin               = "failed to login"

	ErrNotLoggedIn = "not logged in"

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
	loggedIn    bool
}

// DidLogin returns true if the client ever successfully logged in.
func (a *amizoneClient) DidLogin() bool {
	return a.loggedIn
}

// doRequest is an internal proxy method to http.Client.Do which simplifies making requests and handles
// setting custom headers and such. The `method` parameter is the http method to use; endpoint must be relative to
// BaseUrl.
func (a *amizoneClient) doRequest(method string, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, BaseUrl+endpoint, body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", errFailedToComposeRequest, err))
	}
	req.Header.Set("User-Agent", internal.Firefox99UserAgent)
	req.Header.Set("Referer", BaseUrl+"/")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	return a.client.Do(req)
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
	// Amizone uses a "verification" token for logins -- we try to retrieve this from the login form page
	verToken := func() string {
		response, err := a.doRequest(http.MethodGet, "/", nil)
		if err != nil {
			klog.Errorf("%s (login): %s", ErrFailedToVisitPage, err.Error())
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

	loginResponse, err := a.doRequest(http.MethodPost, loginRequestEndpoint, strings.NewReader(loginRequestData.Encode()))
	if err != nil {
		klog.Error("Something went wrong while posting login data: ", err.Error())
		return errors.New(fmt.Sprintf("%s: %s", ErrFailedLogin, err.Error()))
	}

	if loggedIn := parse.LoggedIn(loginResponse.Body); !loggedIn {
		klog.Error("Login failed. Check your credentials.")
		return errors.New(ErrFailedLogin)
	}

	// We need to check if the right tokens are here in the cookie jar to make sure we're logged in
	if !internal.IsLoggedIn(a.client) {
		klog.Error("Failed to login. Are your credentials correct?")
		return errors.New(ErrFailedLogin)
	}

	a.loggedIn = true

	return nil
}

// GetAttendance retrieves, parses and returns attendance data from Amizone
func (a *amizoneClient) GetAttendance() (models.AttendanceRecord, error) {
	response, err := a.doRequest(http.MethodGet, attendancePageEndpoint, nil)
	if err != nil {
		klog.Errorf("%s (attendance): %s", ErrFailedToVisitPage, err.Error())
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
