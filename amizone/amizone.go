package amizone

import (
	"errors"
	"fmt"
	"github.com/ditsuke/go-amizone/amizone/internal"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
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
	examScheduleEndpoint     = "/Examination/ExamSchedule"

	scheduleEndpointTimeFormat = "2006-01-02"

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

// GetAttendance retrieves, parses and returns attendance data from Amizone for courses the client user is enrolled in
// for their latest semester.
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

// GetClassSchedule retrieves, parses and returns class schedule data from Amizone.
// The date parameter is used to determine which schedule to retrieve, but Amizone imposes arbitrary limits on the
// date range, so we have no way of knowing if a request will succeed.
func (a *amizoneClient) GetClassSchedule(date Date) (models.ClassSchedule, error) {
	timeFrom := time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, time.UTC)
	timeTo := timeFrom.Add(time.Hour * 24)

	endpoint := fmt.Sprintf(scheduleEndpointTemplate, timeFrom.Format(scheduleEndpointTimeFormat), timeTo.Format(scheduleEndpointTimeFormat))

	response, err := a.doRequest(true, http.MethodGet, endpoint, nil)
	if err != nil {
		klog.Warningf("request (schedule): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	classSchedule, err := parse.ClassSchedule(response.Body)
	if err != nil {
		klog.Errorf("parse (schedule): %s", err.Error())
		return nil, errors.New(ErrFailedToParsePage)
	}

	return classSchedule, nil
}

// GetExamSchedule retrieves, parses and returns exam schedule data from Amizone.
// Amizone only allows to retrieve the exam schedule for the current semester, and only close to the exam
// dates once the date sheets are out, so we don't take a parameter here.
func (a *amizoneClient) GetExamSchedule() (*models.ExaminationSchedule, error) {
	response, err := a.doRequest(true, http.MethodGet, examScheduleEndpoint, nil)
	if err != nil {
		klog.Warningf("request (exam schedule): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	examSchedule, err := parse.ExaminationSchedule(response.Body)
	if err != nil {
		klog.Errorf("parse (exam schedule): %s", err.Error())
		return nil, errors.New(ErrFailedToParsePage)
	}

	return examSchedule, nil
}
