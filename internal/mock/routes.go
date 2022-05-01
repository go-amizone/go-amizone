package mock

import (
	"errors"
	"fmt"
	"gopkg.in/h2non/gock.v1"
	"net/http"
	"net/url"
	"os"
)

func GockRegisterLoginPage() error {
	mockLogin, err := os.Open("testdata/login_page.html")
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

func GockRegisterLoginRequest(validUsername string, validPassword string) error {
	gock.New("https://s.amizone.net").
		Post("/").
		MatchType("application/x-www-form-urlencoded").
		BodyString(fmt.Sprintf("_Password=%s&_QString=&_UserName=%s&__RequestVerificationToken=.*", url.QueryEscape(validPassword), validUsername)).
		Reply(http.StatusOK).
		AddHeader("Set-Cookie", fmt.Sprintf("ASP.NET_SessionId=%s; path=/; HttpOnly", SessionID)).
		AddHeader("Set-Cookie", fmt.Sprintf("__RequestVerificationToken=%s; path=/; HttpOnly", VerificationToken)).
		AddHeader("Set-Cookie", fmt.Sprintf(".ASPXAUTH=%s; path=/; HttpOnly", AuthCookie))
	return nil
}

func GockRegisterHomePageLoggedIn() error {
	mockHome, err := os.Open("testdata/home_page_logged_in.html")
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
