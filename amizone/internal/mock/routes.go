package mock

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"gopkg.in/h2non/gock.v1"
)

const BaseUrl = "https://s.amizone.net"

// GockRegisterLoginPage registers a gock route for the amizone login page serving the login page from the
// mock filesystem.
func GockRegisterLoginPage() error {
	mockLogin, err := LoginPage.Open()
	if err != nil {
		return errors.New("Failed to open mock login page: " + err.Error())
	}

	gock.New(BaseUrl).
		Get("/").
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockLogin)

	return nil
}

// GockRegisterLoginRequest registers 2 gock routes - one for valid credentials and one for invalid credentials.
// Valid credentials: ValidUser, ValidPass
// Invalid credentials: InvalidUser, InvalidPass
func GockRegisterLoginRequest() error {
	// Valid credentials
	gock.New(BaseUrl).
		Post("/").
		MatchType("application/x-www-form-urlencoded").
		BodyString(fmt.Sprintf("_Password=%s&_QString=&_UserName=%s&__RequestVerificationToken=.*", url.QueryEscape(ValidPass), ValidUser)).
		Reply(http.StatusFound).
		AddHeader("Location", "/Home").
		AddHeader("Set-Cookie", fmt.Sprintf("ASP.NET_SessionId=%s; path=/; HttpOnly", SessionID)).
		AddHeader("Set-Cookie", fmt.Sprintf("__RequestVerificationToken=%s; path=/; HttpOnly", VerificationToken)).
		AddHeader("Set-Cookie", fmt.Sprintf(".ASPXAUTH=%s; path=/; HttpOnly", AuthCookie))

	// 302 redirect to home page on valid credentials
	err := GockRegisterHomePageLoggedIn()
	if err != nil {
		return err
	}

	// Invalid credentials
	gock.New(BaseUrl).
		Post("/").
		MatchType("application/x-www-form-urlencoded").
		BodyString(fmt.Sprintf("_Password=%s&_QString=&_UserName=%s&__RequestVerificationToken=.*", url.QueryEscape(InvalidPass), InvalidUser)).
		Reply(http.StatusFound).
		AddHeader("Location", "/")

	// 302 redirect to login page on invalid credentials
	mockLoginPage, err := LoginPage.Open()
	if err != nil {
		return errors.New("Failed to open mock login page: " + err.Error())
	}
	gock.New(BaseUrl).
		Get("/").
		MatchHeader("Referer", "https://s.amizone.net/").
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockLoginPage)

	return nil
}

// GockRegisterHomePageLoggedIn registers a gock route for the amizone home page, serving the home page for a logged-in
// user from the mock filesystem. The request must have the referrers and cookies expected by the home page.
func GockRegisterHomePageLoggedIn() error {
	mockHome, err := HomePageLoggedIn.Open()
	if err != nil {
		return errors.New("failed to open mock home page: " + err.Error())
	}
	GockRegisterAuthenticatedGet("/Home", mockHome)
	return nil
}

func GockRegisterSemesterCoursesRequest(semesterRef string) error {
	mockCourses, err := CoursesPage.Open()
	if err != nil {
		return errors.New("failed to open mock courses page: " + err.Error())
	}
	GockRegisterAuthenticatedPost("/CourseListSemWise",
		url.Values{"sem": []string{semesterRef}}.Encode(),
		mockCourses,
	)
	return nil
}

func GockRegisterCurrentCoursesPage() error {
	mockCourses, err := CoursesPage.Open()
	if err != nil {
		return errors.New("failed to open mock courses page: " + err.Error())
	}
	GockRegisterAuthenticatedGet("/Academics/MyCourses", mockCourses)
	return nil
}

func GockRegisterSemWiseCoursesPage() error {
	mockCourses, err := CoursesPageSemWise.Open()
	if err != nil {
		return errors.New("failed to open mock courses page: " + err.Error())
	}
	GockRegisterAuthenticatedGet("/Academics/MyCourses", mockCourses)
	return nil
}

// GockRegisterAuthenticatedGet registers an authenticated GET request for the relative endpoint passed.
// The second parameter is used as the response body of the request.
func GockRegisterAuthenticatedGet(endpoint string, responseBody io.Reader) {
	authenticateRequest(newRequest()).
		Get(endpoint).
		Reply(http.StatusOK).
		Type("text/html").
		Body(responseBody)
	return
}

// GockRegisterUnauthenticatedGet registers an unauthenticated GET request for the relative endpoint passed.
func GockRegisterUnauthenticatedGet(endpoint string) error {
	mockLogin, err := LoginPage.Open()
	if err != nil {
		return errors.New("failed to open mock login page: " + err.Error())
	}
	gock.New(BaseUrl).
		Get(endpoint).
		Reply(http.StatusOK).
		Body(mockLogin)

	return nil
}

func GockRegisterAuthenticatedPost(endpoint string, requestBody string, responseBody io.Reader) {
	authenticateRequest(newRequest()).
		Post(endpoint).
		BodyString(requestBody).
		Reply(http.StatusOK).
		Body(responseBody)
	return
}

func GockRegisterUnauthenticatedPost(endpoint string, requestBody string, responseBody io.Reader) {
	newRequest().
		Post(endpoint).
		BodyString(requestBody).
		Reply(http.StatusOK).
		Body(responseBody)
}

func newRequest() *gock.Request {
	return gock.New(BaseUrl).
		MatchHeader("User-Agent", ".*").
		MatchHeader("Referer", BaseUrl)
}

func authenticateRequest(r *gock.Request) *gock.Request {
	return r.
		MatchHeader("Cookie", fmt.Sprintf("ASP.NET_SessionId=%s", SessionID)).
		MatchHeader("Cookie", fmt.Sprintf(".ASPXAUTH=%s", AuthCookie)).
		MatchHeader("Cookie", fmt.Sprintf("__RequestVerificationToken=%s", VerificationToken))
}
