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
	"text/template"
	"time"

	"k8s.io/klog/v2"

	"github.com/ditsuke/go-amizone/amizone/internal"
	"github.com/ditsuke/go-amizone/amizone/internal/marshaller"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	"github.com/ditsuke/go-amizone/amizone/internal/validator"
	"github.com/ditsuke/go-amizone/amizone/models"
)

// Endpoints
const (
	BaseURL = "https://" + internal.AmizoneDomain

	loginRequestEndpoint             = "/"
	attendancePageEndpoint           = "/Home"
	scheduleEndpointTemplate         = "/Calendar/home/GetDiaryEvents?start=%s&end=%s"
	examScheduleEndpoint             = "/Examination/ExamSchedule"
	currentCoursesEndpoint           = "/Academics/MyCourses"
	coursesEndpoint                  = currentCoursesEndpoint + "/CourseListSemWise"
	profileEndpoint                  = "/IDCard"
	macBaseEndpoint                  = "/RegisterForWifi/mac"
	currentExaminationResultEndpoint = "/Examination/Examination"
	examinationResultEndpoint        = currentExaminationResultEndpoint + "/ExaminationListSemWise"
	getWifiMacsEndpoint              = macBaseEndpoint + "/MacRegistration"
	registerWifiMacsEndpoint         = macBaseEndpoint + "/MacRegistrationSave"

	// deleteWifiMacEndpoint is peculiar in that it requires the user's ID as a parameter.
	// This _might_ open doors for an exploit (spoiler: indeed it does)
	removeWifiMacEndpoint = macBaseEndpoint + "/Mac1RegistrationDelete?Amizone_Id=%s&username=%s&X-Requested-With=XMLHttpRequest"

	facultyBaseEndpoint           = "/FacultyFeeback/FacultyFeedback"
	facultyEndpointSubmitEndpoint = facultyBaseEndpoint + "/SaveFeedbackRating"

	atpcPlacementEndpoint = "/Placement/PlacementDetails"
	atpcInternshipEndpoint = atpcPlacementEndpoint + "/IntrenshipIndex";
	atpcCorporateEventEndpoint = "/Placement/CorporatEvent"
)

// Miscellaneous
const (
	classScheduleEndpointDateFormat = "2006-01-02"

	verificationTokenName = "__RequestVerificationToken"
)

// Errors
const (
	ErrBadClient              = "the http client passed must have a cookie jar, or be nil"
	ErrFailedToVisitPage      = "failed to visit page"
	ErrFailedToFetchPage      = "failed to fetch page"
	ErrFailedToReadResponse   = "failed to read response body"
	ErrFailedLogin            = "failed to login"
	ErrInvalidCredentials     = ErrFailedLogin + ": invalid credentials"
	ErrInternalFailure        = "internal failure"
	ErrFailedToComposeRequest = ErrInternalFailure + ": failed to compose request"
	ErrFailedToParsePage      = ErrInternalFailure + ": failed to parse page"
	ErrInvalidMac             = "invalid MAC address passed"
	ErrNoMacSlots             = "no free wifi mac slots"
	ErrFailedToRegisterMac    = "failed to register mac address"
)

type Credentials struct {
	Username string
	Password string
}

// Client is the main struct for the amizone package, exposing the entire API surface
// for the portal as implemented here. The struct must always be initialized through a public
// constructor like NewClient()
type Client struct {
	httpClient  *http.Client
	credentials *Credentials
	// muLogin is a mutex that protects the lastAttempt and didLogin fields from concurrent access.
	muLogin struct {
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
			klog.Error("failed to create cookiejar for the amizone client. this is a bug.")
			return nil, errors.New(ErrInternalFailure)
		}
		httpClient = &http.Client{Jar: jar}
	}

	if jar := httpClient.Jar; jar == nil {
		klog.Error("amizone.NewClient called with a jar-less http client. please pass a client with a non-nil cookie jar")
		return nil, errors.New(ErrBadClient)
	}

	client := &Client{
		httpClient:  httpClient,
		credentials: &cred,
	}

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

	// Record our last login attempt so that we can avoid trying again for some time.
	a.muLogin.lastAttempt = time.Now()

	// Amizone uses a "verification" token for logins -- we try to retrieve this from the login form page
	getVerificationTokenFromLoginPage := func() string {
		response, err := a.doRequest(false, http.MethodGet, "/", nil)
		if err != nil {
			klog.Errorf("login: %s", err.Error())
			return ""
		}
		return parse.VerificationToken(response.Body)
	}()

	if getVerificationTokenFromLoginPage == "" {
		klog.Error("login: failed to retrieve verification token from the login page")
		return fmt.Errorf("%s: %s", ErrFailedLogin, ErrFailedToParsePage)
	}

	loginRequestData := func() (v url.Values) {
		v = url.Values{}
		v.Set(verificationTokenName, getVerificationTokenFromLoginPage)
		v.Set("_UserName", a.credentials.Username)
		v.Set("_Password", a.credentials.Password)
		v.Set("_QString", "")
		return
	}()

	loginResponse, err := a.doRequest(
		false,
		http.MethodPost,
		loginRequestEndpoint,
		strings.NewReader(loginRequestData.Encode()),
	)
	if err != nil {
		klog.Warningf("error while making HTTP request to the amizone login page: %s", err.Error())
		return fmt.Errorf("%s: %w", ErrFailedLogin, err)
	}

	// The login request should redirect our request to the home page with a 302 "found" status code.
	// If we're instead redirected to the login page, we've failed to log in because of invalid credentials
	if loginResponse.Request.URL.Path == loginRequestEndpoint {
		return errors.New(ErrInvalidCredentials)
	}

	if loggedIn := parse.IsLoggedIn(loginResponse.Body); !loggedIn {
		klog.Error(
			"login attempt failed as indicated by parsing the page returned after the login request, while the redirect indicated that it passed." +
				" this failure indicates that something broke between Amizone and go-amizone.",
		)
		return errors.New(ErrFailedLogin)
	}

	if !internal.IsLoggedIn(a.httpClient) {
		klog.Error(
			"login attempt failed as indicated by checking the cookies in the http client's cookie jar. this failure indicates that something has broken between" +
				" Amizone and go-amizone, possibly the cookies used by amizone for authentication.",
		)
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
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	attendanceRecord, err := parse.Attendance(response.Body)
	if err != nil {
		klog.Errorf("parse (attendance): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return models.AttendanceRecords(attendanceRecord), nil
}

// GetExaminationResult retrieves, parses and returns a ExaminationResultRecords from Amizone for their latest semester
// for which the result is available
func (a *Client) GetCurrentExaminationResult() (*models.ExamResultRecords, error) {
	response, err := a.doRequest(true, http.MethodGet, currentExaminationResultEndpoint, nil)
	if err != nil {
		klog.Warningf("request (examination-result): %s", err.Error())
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	examinationResultRecords, err := parse.ExaminationResult(response.Body)
	if err != nil {
		klog.Errorf("parse (examination-result): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return examinationResultRecords, nil
}

// GetExaminationResult retrieves, parses and returns a ExaminationResultRecords from Amizone for the semester referred by
// semesterRef. Semester references should be retrieved through GetSemesters, which returns a list of valid
// semesters with names and references.
func (a *Client) GetExaminationResult(semesterRef string) (*models.ExamResultRecords, error) {
	payload := url.Values{
		"sem": []string{semesterRef},
	}.Encode()

	response, err := a.doRequest(true, http.MethodPost, examinationResultEndpoint, strings.NewReader(payload))
	if err != nil {
		klog.Warningf("request (examination-result): %s", err.Error())
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	examinationResultRecords, err := parse.ExaminationResult(response.Body)
	if err != nil {
		klog.Errorf("parse (examination-result): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return examinationResultRecords, nil
}

// GetClassSchedule retrieves, parses and returns class schedule data from Amizone.
// The date parameter is used to determine which schedule to retrieve, however as Amizone imposes arbitrary limits on the
// date range, as in scheduled for dates older than some months are not stored by Amizone, we have no way of knowing if a request will succeed.
func (a *Client) GetClassSchedule(year int, month time.Month, date int) (models.ClassSchedule, error) {
	timeFrom := time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
	timeTo := timeFrom.Add(time.Hour * 24)

	endpoint := fmt.Sprintf(
		scheduleEndpointTemplate,
		timeFrom.Format(classScheduleEndpointDateFormat),
		timeTo.Format(classScheduleEndpointDateFormat),
	)

	response, err := a.doRequest(true, http.MethodGet, endpoint, nil)
	if err != nil {
		klog.Warningf("request (schedule): %s", err.Error())
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	classSchedule, err := parse.ClassSchedule(response.Body)
	if err != nil {
		klog.Errorf("parse (schedule): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrFailedToParsePage, err)
	}
	// Filter classes by start date, since might also return classes for the dates before/after the target date.
	scheduledClassesForTargetDate := classSchedule.FilterByDate(timeFrom)

	return models.ClassSchedule(scheduledClassesForTargetDate), nil
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
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
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
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	courses, err := parse.Courses(response.Body)
	if err != nil {
		klog.Errorf("parse (current courses): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return models.Courses(courses), nil
}

// GetAtpcListings retrieves, parses and returns the ATPC placement, internship, and corporate event details from Amizone.
func (a *Client) GetAtpcListings() (*models.AtpcListings, error) {
	
	corporateEventResponse, err := a.doRequest(true, http.MethodGet, atpcCorporateEventEndpoint, nil)
	if err != nil {
		klog.Warningf("request (atpc corporate event): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	corporateEventDetails, err := parse.AtpcDetails(corporateEventResponse.Body)
	if err != nil {
		klog.Errorf("parse (atpc corporate event): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}
	placementResponse, err := a.doRequest(true, http.MethodGet, atpcPlacementEndpoint, nil)
	if err != nil {
		klog.Warningf("request (atpc placement): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	placementDetails, err := parse.AtpcDetails(placementResponse.Body)
	if err != nil {
		klog.Errorf("parse (atpc placement): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	internshipResponse, err := a.doRequest(true, http.MethodGet, atpcInternshipEndpoint, nil)
	if err != nil {
		klog.Warningf("request (atpc internship): %s", err.Error())
		return nil, errors.New(ErrFailedToVisitPage)
	}

	internshipDetails, err := parse.AtpcDetails(internshipResponse.Body)
	if err != nil {
		klog.Errorf("parse (atpc internship): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}
	

	details := models.AtpcListings{
		Placement: placementDetails,
		Internship: internshipDetails,
		CorporateEvent: corporateEventDetails,
	}

	return &details, nil
}

// GetUserProfile retrieves, parsed and returns the current user's profile from Amizone.
func (a *Client) GetUserProfile() (*models.Profile, error) {
	response, err := a.doRequest(true, http.MethodGet, profileEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get profile): %s", err.Error())
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	profile, err := parse.Profile(response.Body)
	if err != nil {
		klog.Errorf("parse (profile): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return (*models.Profile)(profile), nil
}

func (a *Client) GetWiFiMacInformation() (*models.WifiMacInfo, error) {
	response, err := a.doRequest(true, http.MethodGet, getWifiMacsEndpoint, nil)
	if err != nil {
		klog.Warningf("request (get wifi macs): %s", err.Error())
		return nil, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	info, err := parse.WifiMacInfo(response.Body)
	if err != nil {
		klog.Errorf("parse (wifi macs): %s", err.Error())
		return nil, fmt.Errorf("%s: %w", ErrInternalFailure, err)
	}

	return (*models.WifiMacInfo)(info), nil
}

// RegisterWifiMac registers a mac address on Amizone.
// If bypassLimit is true, it bypasses Amizone's artificial 2-address
// limitation. However, only the 2 oldest mac addresses are reflected
// in the GetWifiMacInfo response.
// TODO: is the bypassLimit functional?
func (a *Client) RegisterWifiMac(addr net.HardwareAddr, bypassLimit bool) error {
	// validate
	err := validator.ValidateHardwareAddr(addr)
	if err != nil {
		return errors.New(ErrInvalidMac)
	}
	wifiInfo, err := a.GetWiFiMacInformation()
	if err != nil {
		klog.Warningf("failure while getting wifi mac info: %s", err.Error())
		return err
	}

	if wifiInfo.IsRegistered(addr) {
		klog.Infof("wifi already registered.. skipping request")
		return nil
	}

	if !wifiInfo.HasFreeSlot() {
		if !bypassLimit {
			return errors.New(ErrNoMacSlots)
		}
		// Remove the last mac address :)
		wifiInfo.RegisteredAddresses = wifiInfo.RegisteredAddresses[:len(wifiInfo.RegisteredAddresses)-1]
	}

	wifis := append(wifiInfo.RegisteredAddresses, addr)

	payload := url.Values{}
	payload.Set(verificationTokenName, wifiInfo.GetRequestVerificationToken())
	// ! VULN: register mac as anyone or no one by changing this ID.
	payload.Set("Amizone_Id", a.credentials.Username)

	// _Name_ is a dummy field, as in it doesn't matter what its value is, but it needs to be present.
	// I suspect this might go straight into the DB.
	payload.Set("Name", "DoesntMatter")

	for i, mac := range wifis {
		payload.Set(fmt.Sprintf("Mac%d", i+1), marshaller.Mac(mac))
	}

	res, err := a.doRequest(true, http.MethodPost, registerWifiMacsEndpoint, strings.NewReader(payload.Encode()))
	if err != nil {
		klog.Errorf("request (register wifi mac): %s", err.Error())
		return fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}
	// We attempt to verify if the mac was set successfully, but its futile if bypassLimit was used since Amizone only exposes
	if bypassLimit {
		return nil
	}

	macs, err := parse.WifiMacInfo(res.Body)
	if err != nil {
		klog.Errorf("parse (wifi macs): %s", err.Error())
		return errors.New(ErrFailedToParsePage)
	}
	if !macs.IsRegistered(addr) {
		klog.Errorf("mac not registered: %s", addr.String())
		return errors.New(ErrFailedToRegisterMac)
	}

	return nil
}

// RemoveWifiMac removes a mac address from the Amizone mac address registry. If the mac address is not registered in the
// first place, this function does nothing.
func (a *Client) RemoveWifiMac(addr net.HardwareAddr) error {
	err := validator.ValidateHardwareAddr(addr)
	if err != nil {
		return errors.New(ErrInvalidMac)
	}

	// ! VULN: remove mac addresses registered by anyone if you know the mac/username pair.
	response, err := a.doRequest(
		true,
		http.MethodGet,
		fmt.Sprintf(removeWifiMacEndpoint, a.credentials.Username, marshaller.Mac(addr)),
		nil,
	)
	if err != nil {
		klog.Errorf("request (remove wifi mac): %s", err.Error())
		return fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	wifiInfo, err := parse.WifiMacInfo(response.Body)
	if err != nil {
		klog.Errorf("parse (wifi macs): %s", err.Error())
		return errors.New(ErrFailedToParsePage)
	}

	if wifiInfo.IsRegistered(addr) {
		return errors.New("failed to remove mac address")
	}

	return nil
}

// SubmitFacultyFeedbackHack submits feedback for *all* faculties, giving the same ratings and comments to all.
// This is a hack because we're not allowing fine-grained control over feedback points or individual faculties. This is
// because the form is a pain to parse, and the feedback system is a pain to work with in general.
// Returns: the number of faculties for which feedback was submitted. Note that this number would be zero
// if the feedback was already submitted or is not open.
func (a *Client) SubmitFacultyFeedbackHack(rating int32, queryRating int32, comment string) (int32, error) {
	// Validate
	if rating > 5 || rating < 1 {
		return 0, errors.New("invalid rating")
	}
	if queryRating > 3 || queryRating < 1 {
		return 0, errors.New("invalid query rating")
	}
	if comment == "" {
		return 0, errors.New("comment cannot be empty")
	}

	// Transform queryRating for "higher number is higher rating" semantics (it's the opposite in the form 😭)
	switch queryRating {
	case 1:
		queryRating = 3
	case 3:
		queryRating = 1
	}

	facultyPage, err := a.doRequest(true, http.MethodGet, facultyBaseEndpoint, nil)
	if err != nil {
		klog.Errorf("request (faculty page): %s", err.Error())
		return 0, fmt.Errorf("%s: %s", ErrFailedToFetchPage, err.Error())
	}

	feedbackSpecs, err := parse.FacultyFeedback(facultyPage.Body)
	if err != nil {
		klog.Errorf("parse (faculty feedback): %s", err.Error())
		return 0, errors.New(ErrFailedToParsePage)
	}

	payloadTemplate, err := template.New("facultyFeedback").Parse(facultyFeedbackTpl)
	if err != nil {
		klog.Errorf("Error parsing faculty feedback template: %s", err.Error())
		return 0, errors.New(ErrInternalFailure)
	}

	// Parallelize feedback submission for max gains 📈
	wg := sync.WaitGroup{}
	for _, spec := range feedbackSpecs {
		spec.Set__Rating = fmt.Sprint(rating)
		spec.Set__Comment = url.QueryEscape(comment)
		spec.Set__QRating = fmt.Sprint(queryRating)

		payloadBuilder := strings.Builder{}
		err = payloadTemplate.Execute(&payloadBuilder, spec)
		if err != nil {
			klog.Errorf("Error executing faculty feedback template: %s", err.Error())
			return 0, fmt.Errorf("error marshalling feedback request: %s", err)
		}
		wg.Add(1)
		go func(payload string) {
			response, err := a.doRequest(true, http.MethodPost, facultyEndpointSubmitEndpoint, strings.NewReader(payload))
			if err != nil {
				klog.Errorf("error submitting a faculty feedback: %s", err.Error())
			}
			if response.StatusCode != http.StatusOK {
				klog.Errorf("Unexpected non-200 status code from faculty feedback submission: %d", response.StatusCode)
			}
			wg.Done()
		}(payloadBuilder.String())
	}

	wg.Wait()
	return int32(len(feedbackSpecs)), nil
}
