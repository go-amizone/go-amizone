package amizone

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ditsuke/go-amizone/amizone/internal"
	"github.com/ditsuke/go-amizone/amizone/internal/marshaller"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/models"
	"k8s.io/klog/v2"
)

// Endpoints
const (
	BaseUrl = "https://" + internal.AmizoneDomain

	loginRequestEndpoint     = "/"
	attendancePageEndpoint   = "/Home"
	scheduleEndpointTemplate = "/Calendar/home/GetDiaryEvents?start=%s&end=%s"
	examScheduleEndpoint     = "/Examination/ExamSchedule"
	currentCoursesEndpoint   = "/Academics/MyCourses"
	coursesEndpoint          = currentCoursesEndpoint + "/CourseListSemWise"
	profileEndpoint          = "/IDCard"
	macBaseEndpoint          = "/RegisterForWifi/mac"
	getWifiMacsEndpoint      = macBaseEndpoint + "/MacRegistration"
	registerWifiMacsEndpoint = macBaseEndpoint + "/MacRegistrationSave"

	// deleteWifiMacEndpoint is peculiar in that it requires the user's ID as a parameter.
	// This _might_ open doors for an exploit (spoiler: indeed it does)
	removeWifiMacEndpoint = macBaseEndpoint + "/Mac1RegistrationDelete?username=%s&Amizone_Id=%s"
)

// Miscellaneous
const (
	scheduleEndpointTimeFormat = "2006-01-02"

	verificationTokenName = "__RequestVerificationToken"
)

// Errors
const (
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

// Client is the main struct for the amizone package, exposing the entire API surface
// for the portal as implemented here. The struct must always be initialised through a public
// constructor like NewClient()
type Client struct {
	client      *http.Client
	credentials *Credentials
	muLogin     struct {
		sync.Mutex
		lastAttempt time.Time
		didLogin    bool
	}
}

// DidLogin returns true if the client ever successfully logged in.
func (a *Client) DidLogin() bool {
	a.muLogin.Lock()
	defer a.muLogin.Unlock()
	return a.muLogin.didLogin
}

// NewClient create a new client instance with Credentials passed, then attempts to log in to the website.
// The *http.Client parameter can be nil, in which case a default client will be created in its place.
// To get a non-logged in client, pass empty credentials, ala Credentials{}.
func NewClient(cred Credentials, httpClient *http.Client) (*Client, error) {
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

	client := &Client{
		client:      httpClient,
		credentials: &cred,
	}

	// We don't try to log in if empty credentials were passed
	if cred == (Credentials{}) {
		return client, nil
	}

	return client, client.login()
}

// login attempts to log in to Amizone with the credentials passed to the Client and a scrapped
// "__RequestVerificationToken" value.
func (a *Client) login() error {
	a.muLogin.Lock()
	defer a.muLogin.Unlock()

	if time.Since(a.muLogin.lastAttempt) < time.Minute*2 {
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
		return fmt.Errorf("%s: %s", ErrFailedLogin, ErrFailedToParsePage)
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
		return fmt.Errorf("%s: %w", ErrFailedLogin, err)
	}

	// The login request should redirect our request to the home page with a 302 "found" status code.
	// If we're instead redirected to the login page, we've failed to log in because of invalid credentials
	if loginResponse.Request.URL.Path == loginRequestEndpoint {
		return errors.New(ErrInvalidCredentials)
	}

	if loggedIn := parse.IsLoggedIn(loginResponse.Body); !loggedIn {
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
func (a *Client) GetAttendance() (models.AttendanceRecords, error) {
	response, err := a.doRequest(true, http.MethodGet, attendancePageEndpoint, nil)
	if err != nil {
		klog.Warningf("request (attendance): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	attendanceRecord, err := parse.Attendance(response.Body)
	if err != nil {
		klog.Errorf("parse (attendance): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return models.AttendanceRecords(attendanceRecord), nil
}

// Getmodels.ClassSchedule retrieves, parses and returns class schedule data from Amizone.
// The date parameter is used to determine which schedule to retrieve, but Amizone imposes arbitrary limits on the
// date range, so we have no way of knowing if a request will succeed.
func (a *Client) ClassSchedule(year int, month time.Month, date int) (models.ClassSchedule, error) {
	timeFrom := time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
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
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}
	filteredSchedule := classSchedule.FilterByDate(timeFrom)

	return models.ClassSchedule(filteredSchedule), nil
}

// GetExamSchedule retrieves, parses and returns exam schedule data from Amizone.
// Amizone only allows to retrieve the exam schedule for the current semester, and only close to the exam
// dates once the date sheets are out, so we don't take a parameter here.
func (a *Client) GetExamSchedule() (*models.ExaminationSchedule, error) {
	response, err := a.doRequest(true, http.MethodGet, examScheduleEndpoint, nil)
	if err != nil {
		klog.Warningf("request (exam schedule): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	examSchedule, err := parse.ExaminationSchedule(response.Body)
	if err != nil {
		klog.Errorf("parse (exam schedule): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return (*models.ExaminationSchedule)(examSchedule), nil
}

// GetSemesters retrieves, parses and returns a SemesterList from Amizone. This list includes all semesters for which
// information can be retrieved through other semester-specific methods like GetCourses.
func (a *Client) GetSemesters() (models.SemesterList, error) {
	response, err := a.doRequest(true, http.MethodGet, currentCoursesEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get semesters): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	semesters, err := parse.Semesters(response.Body)
	if err != nil {
		klog.Errorf("parse (semesters): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return (models.SemesterList)(semesters), nil
}

// GetCourses retrieves, parses and returns a SemesterList from Amizone for the semester referred by
// semesterRef. Semester references should be retrieved through GetSemesters, which returns a list of valid
// semesters with names and references.
func (a *Client) GetCourses(semesterRef string) (models.Courses, error) {
	payload := url.Values{
		"sem": []string{semesterRef},
	}.Encode()

	response, err := a.doRequest(true, http.MethodPost, coursesEndpoint, strings.NewReader(payload))
	if err != nil {
		klog.Warningf("request (get courses): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	courses, err := parse.Courses(response.Body)
	if err != nil {
		klog.Errorf("parse (courses): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return models.Courses(courses), nil
}

// GetCurrentCourses retrieves, parses and returns a SemesterList from Amizone for the most recent semester.
func (a *Client) GetCurrentCourses() (models.Courses, error) {
	response, err := a.doRequest(true, http.MethodGet, currentCoursesEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get current courses): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	courses, err := parse.Courses(response.Body)
	if err != nil {
		klog.Errorf("parse (current courses): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return models.Courses(courses), nil
}

// GetProfile retrieves, parsed and returns the current user's profile from Amizone.
func (a *Client) GetProfile() (*models.Profile, error) {
	response, err := a.doRequest(true, http.MethodGet, profileEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get profile): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	profile, err := parse.Profile(response.Body)
	if err != nil {
		klog.Errorf("parse (profile): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return (*models.Profile)(profile), nil
}

func (a *Client) GetWifiMacInfo() (*models.WifiMacInfo, error) {
	response, err := a.doRequest(true, http.MethodGet, getWifiMacsEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get wifi macs): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	info, err := parse.WifiMacInfo(response.Body)
	if err != nil {
		klog.Errorf("parse (wifi macs): %s", err.Error())
		return nil, errors.New(ErrFailedToParsePage)
	}

	return (*models.WifiMacInfo)(info), nil
}

// RegisterWifiMac registers a mac address on Amizone.
// If bypassLimit is true, it bypasses Amizone's artificial 2-address
// limitation. However, only the 2 oldest mac addresses are reflected
// in the GetWifiMacInfo response.
// TODO: is the overwriteExisting functional?
func (a *Client) RegisterWifiMac(addr net.HardwareAddr, bypassLimit bool) error {
	info, err := a.GetWifiMacInfo()
	if err != nil {
		klog.Warningf("failures while getting wifi mac info: %s", err.Error())
		return err
	}

	if !info.HasFreeSlot() {
		// but the limitation is artificial so... we do nothing?
		// we shouldn't be defaulting to the bypass-style behaviour, though
		// TODO: flag or param to enable the bypass behavior
		if !bypassLimit {
			return errors.New("no free wifi slots")
		}
		// Remove the last mac address :)
		info.RegisteredAddresses = info.RegisteredAddresses[:len(info.RegisteredAddresses)-1]
	}

	if info.IsRegistered(addr) {
		klog.Infof("wifi already registered.. skipping request")
		return nil
	}

	wifis := append(info.RegisteredAddresses, addr)

	payload := url.Values{}
	payload.Set(verificationTokenName, info.GetRequestVerificationToken())
	// ! VULN: register mac as anyone or no one by changing this ID.
	payload.Set("Amizone_Id", a.credentials.Username)

	// _Name_ is a dummy field, as in it doesn't matter what its value is, but it needs to be present.
	// I suspect this might go straight into the DB.
	payload.Set("Name", "DoesntMatter")

	payload.Set("Mac1", marshaller.Mac(wifis[0]))
	payload.Set("Mac2", func() string {
		if len(wifis) == 2 {
			return marshaller.Mac(wifis[1])
		}
		return ""
	}())
	if len(wifis) == 2 {
		payload.Set("Mac2", marshaller.Mac(wifis[1]))
	}

	// here we make a POST form submission to the form
	// open question: does the endpoint necessarility need extranneous userinfo (name, admission number (probably yes for this one))

	// Open question: _should_ we be verifying the response? We _could_ parse out the updated mac list and verify that it has our new mac,
	// but the failure modes are many and the only thing we can do (as of now) is move on. Especially since we're already verifying the
	// validity of the mac addresses before we even enter this function.
	_, err = a.doRequest(true, http.MethodPost, registerWifiMacsEndpoint, strings.NewReader(payload.Encode()))
	if err != nil {
		klog.Errorf("request (register wifi mac): %s", err.Error())
		return errors.New(ErrFailedToVisitPage)
	}

	return nil
}

// RemoveWifiMac removes a mac address from the Amizone mac address registry. If the mac address is not registered in the
// first place, this function does nothing.
func (a *Client) RemoveWifiMac(addr string) error {
	// just make the GET request here to delete the mac
	mac, err := net.ParseMAC(addr)
	if err != nil {
		return errors.New("invalid mac address")
	}

	// ! VULN: remove mac addresses registered by anyone if you know the mac/username pair.
	response, err := a.doRequest(true, http.MethodGet, fmt.Sprintf(removeWifiMacEndpoint, a.credentials.Username, marshaller.Mac(mac)), nil)
	if err != nil {
		klog.Errorf("request (remove wifi mac): %s", err.Error())
		return errors.New(ErrFailedToVisitPage)
	}

	wifiInfo, err := parse.WifiMacInfo(response.Body)
	if err != nil {
		klog.Errorf("parse (wifi macs): %s", err.Error())
		return errors.New(ErrFailedToParsePage)
	}

	if wifiInfo.IsRegistered(mac) {
		return errors.New("failed to remove mac address")
	}

	return nil
}
