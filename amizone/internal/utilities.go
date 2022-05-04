package internal

import (
	"net/http"
	"net/url"
)

type cookieMap map[string]string

func (c cookieMap) contains(key string) bool {
	if _, ok := c[key]; ok {
		return true
	}
	return false
}

// IsLoggedIn returns true if the amizone client has the cookies to be logged in.
// This method does not check if the cookies are still valid.
func IsLoggedIn(client *http.Client) bool {
	jar := client.Jar
	if jar == nil {
		return false
	}

	amizoneUrl, _ := url.Parse("https://" + AmizoneDomain)

	amizoneCookies := func() cookieMap {
		cookieMap := make(cookieMap)
		for _, cookie := range jar.Cookies(amizoneUrl) {
			cookieMap[cookie.Name] = cookie.Value
		}
		return cookieMap
	}()

	for _, key := range []string{".ASPXAUTH", "ASP.NET_SessionId", "__RequestVerificationToken"} {
		if !amizoneCookies.contains(key) {
			return false
		}
	}

	return true
}
