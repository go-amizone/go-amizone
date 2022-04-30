package amizone

import (
	"GoFriday/lib/amizone/internal"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

const (
	BaseUrl                  = "https://s.amizone.net"
	attendancePageEndpoint   = "/Home"
	scheduleEndpointTemplate = "/Calendar/home/GetDiaryEvents?start=%s&end=%s"

	scheduleEndpointTimeFormat = "2006-01-02"
	scheduleJsonTimeFormat     = "2006/01/02 03:04:05 PM"

	verificationTokenName = "__RequestVerificationToken"

	ErrFailedAttendanceRetrieval = "failed to retrieve attendance"
	ErrFailedToVisitPage         = "failed to visit page"
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
	c := internal.GetNewColly(a.client, false)

	// Amizone uses a "verification" token for logins -- we try to retrieve this from the login form page
	var verToken string

	c.OnHTML(fmt.Sprintf("input[name='%s']", verificationTokenName), func(e *colly.HTMLElement) {
		verToken = e.Attr("value")
	})

	c.OnResponse(func(r *colly.Response) {
		klog.Infof("Receiving response from amizone with status: %d", r.StatusCode)
	})

	if err := c.Visit(BaseUrl); err != nil {
		klog.Error("Something went wrong while visiting the login page: " + err.Error())
		return errors.New("failed to visit login page: " + err.Error())
	}

	if verToken == "" {
		klog.Error("Failed to retrieve verification token from login page. What's up?")
		return errors.New("could not find verification token")
	}

	loginRequestData := map[string]string{
		"__RequestVerificationToken": verToken,
		"_UserName":                  a.credentials.Username,
		"_Password":                  a.credentials.Password,
		"_QString":                   "",
	}

	if err := c.Post(BaseUrl, loginRequestData); err != nil {
		klog.Error("Something went wrong while posting login data: ", err.Error())
		return errors.New("could not post login data: " + err.Error())
	}

	// We need to check if the right tokens are here in the cookie jar to make sure we're logged in
	if !internal.IsLoggedIn(a.client) {
		klog.Error("Failed to login. Are your credentials correct?")
		return errors.New("failed to login")
	}
	return nil
}

// GetAttendance retrieves the attendanceRecord from Amizone
func (a *amizoneClient) GetAttendance() (AttendanceRecord, error) {
	res := make(AttendanceRecord)
	var recordListFound bool

	c := internal.GetNewColly(a.client, true)

	c.OnHTML("#tasks", func(e *colly.HTMLElement) {
		recordListFound = true
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			course := &course{
				code: el.ChildText("span.sub-code"),
				name: func() string {
					rawInner := el.ChildText("span.lbl")
					spaceIndex := strings.IndexRune(rawInner, ' ')
					return strings.TrimSpace(rawInner[spaceIndex:])
				}(),
			}

			if course.code == "" {
				klog.Warning("Failed to parse course code for an attendance list item")
				return
			}

			attendance := func() *courseAttendance {
				raw := el.ChildText("div.class-count span")
				divided := strings.Split(raw, "/")
				if len(divided) != 2 {
					klog.Warning("Attendance string has unexpected format!", course.code)
					return nil
				}
				return &courseAttendance{
					course:          course,
					classesAttended: divided[0],
					classesHeld:     divided[1],
				}
			}()

			if attendance == nil {
				klog.Warningf("Failed to parse attendance for course: %s", course.code)
				return
			}

			res[course.code] = attendance
		})
	})

	if err := c.Visit(BaseUrl + attendancePageEndpoint); err != nil {
		klog.Error("Something went wrong while visiting the attendance page: " + err.Error())
		return nil, errors.New("failed to visit the attendance page: " + err.Error())
	}

	if !recordListFound {
		klog.Error("Failed to find the attendance list on the attendance page. Did we login at all?")
		return nil, errors.New(ErrFailedAttendanceRetrieval)
	}

	return res, nil
}

func (a *amizoneClient) GetClassSchedule(date Date) (classSchedule, error) {
	timeFrom := time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, time.UTC)
	timeTo := timeFrom.Add(time.Hour * 24)

	var schedule classSchedule
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

			class := &scheduledClass{
				course: &course{
					code: entry.CourseCode,
					name: entry.CourseName,
				},
				startTime: timeParserFunc(entry.Start),
				endTime:   timeParserFunc(entry.End),
				faculty:   entry.Faculty,
				room:      entry.Room,
			}

			schedule = append(schedule, class)
		}
	})

	err := c.Visit(BaseUrl + fmt.Sprintf(scheduleEndpointTemplate, timeFrom.Format(scheduleEndpointTimeFormat), timeTo.Format(scheduleEndpointTimeFormat)))
	if err != nil {
		klog.Error("Something went wrong while visiting the schedule endpoint: " + err.Error())
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToVisitPage, err.Error()))
	}

	// @todo: We might want to sort the schedule here to ensure that the classes are in the right order

	return schedule, nil
}
