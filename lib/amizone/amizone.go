package amizone

import (
	"errors"
	"fmt"
	"github.com/gocolly/colly/v2"
	"k8s.io/klog/v2"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

const (
	baseUrl                = "https://s.amizone.net"
	attendancePageEndpoint = "/Home"
	verificationTokenName  = "__RequestVerificationToken"
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

// NewClient create a new amizoneClient instance with Credentials passed, then attempts to log in to the website.
// The *http.Client parameter can be nil, in which case a default client
// will be created in its place.
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
	err := client.login()

	return client, err
}

// login attempts to log in to Amizone with the credentials passed to the amizoneClient and a scrapped
// "__RequestVerificationToken" value.
func (a *amizoneClient) login() error {
	c := getNewColly(a.client, false)

	// Amizone uses a "verification" token for logins -- we try to retrieve this from the login form page
	var verToken string

	c.OnHTML(fmt.Sprintf("input[name='%s']", verificationTokenName), func(e *colly.HTMLElement) {
		verToken = e.Attr("value")
	})

	c.OnResponse(func(r *colly.Response) {
		klog.Infof("Receiving response from amizone with status: %d", r.StatusCode)
	})

	if err := c.Visit(baseUrl); err != nil {
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

	if err := c.Post(baseUrl, loginRequestData); err != nil {
		klog.Error("Something went wrong while posting login data: ", err.Error())
		return errors.New("could not post login data: " + err.Error())
	}

	// We need to check if the right tokens are here in the cookie jar to make sure we're logged in
	if !isLoggedIn(a.client) {
		klog.Error("Failed to login. Are your credentials correct?")
		return errors.New("failed to login")
	}
	return nil
}

// GetAttendance retrieves the attendanceRecord from Amizone
func (a *amizoneClient) GetAttendance() (attendanceRecord, error) {
	res := make(attendanceRecord)
	var recordListFound bool

	c := getNewColly(a.client, true)

	c.OnHTML("#tasks", func(e *colly.HTMLElement) {
		recordListFound = true
		e.ForEach("li", func(_ int, el *colly.HTMLElement) {
			course := &course{
				code: el.ChildText("span.sub-code"),
				name: el.ChildText("span.lbl"),
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
					classesAttended: divided[0],
					classesHeld:     divided[1],
				}
			}()

			if attendance == nil {
				klog.Warningf("Failed to parse attendance for course: %s", course.code)
				return
			}

			res[course] = attendance
		})
	})

	if err := c.Visit(baseUrl + attendancePageEndpoint); err != nil {
		klog.Error("Something went wrong while visiting the attendance page: " + err.Error())
		return nil, err
	}

	if !recordListFound {
		klog.Error("Failed to find the attendance list on the attendance page. Did we login at all?")
		return nil, errors.New("could not find attendance list")
	}

	return res, nil
}
