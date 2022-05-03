package mock

import (
	"errors"
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"net/url"
)

func GockRegisterLoginPage() error {
	mockLogin, err := FS.Open(LoginPage)
	if err != nil {
		return errors.New("Failed to open mock login page: " + err.Error())
	}

	gock.New("https://s.amizone.net").
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
	gock.New("https://s.amizone.net").
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
	gock.New("https://s.amizone.net").
		Post("/").
		MatchType("application/x-www-form-urlencoded").
		BodyString(fmt.Sprintf("_Password=%s&_QString=&_UserName=%s&__RequestVerificationToken=.*", url.QueryEscape(InvalidPass), InvalidUser)).
		Reply(http.StatusFound).
		AddHeader("Location", "/")

	// 302 redirect to login page on invalid credentials
	mockLoginPage, err := FS.Open(LoginPage)
	if err != nil {
		return errors.New("Failed to open mock login page: " + err.Error())
	}
	gock.New("https://s.amizone.net").
		Get("/").
		MatchHeader("Referer", "https://s.amizone.net/").
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockLoginPage)

	return nil
}

func GockRegisterHomePageLoggedIn() error {
	mockHome, err := FS.Open(HomePageLoggedIn)
	if err != nil {
		return errors.New("Failed to open mock home page: " + err.Error())
	}

	gock.New("https://s.amizone.net").
		Get("/Home").
		MatchHeader("User-Agent", ".*").
		MatchHeader("Referer", "https://s.amizone.net").
		MatchHeader("Cookie", fmt.Sprintf("ASP.NET_SessionId=%s", SessionID)).
		MatchHeader("Cookie", fmt.Sprintf(".ASPXAUTH=%s", AuthCookie)).
		MatchHeader("Cookie", fmt.Sprintf("__RequestVerificationToken=%s", VerificationToken)).
		Reply(http.StatusOK).
		Type("text/html").
		Body(mockHome)
	return nil
}
