package parse

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

const loginFormHtmlId = "loginform"

// IsLoggedIn attempts to determine whether a response body indicates an authenticated session.
// To achieve this, this function will first attempt to parse the body as an HTML document, failing to do
// which is assumed to indicate an authenticated session because Amizone seems to redirect unauthenticated requests
// from all endpoints to the login page.
// If the body is parsed into a HTMl document, this function will attempt to find the login form; failing to find
// the login form is assumed to indicate an authenticated session.
func IsLoggedIn(body io.Reader) bool {
	// Try to find the login form
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil { // Failure to parse an HTML document ~ logged-in
		return true
	}
	return IsLoggedInDOM(doc)
}

func IsLoggedInDOM(doc *goquery.Document) bool {
	loginFormMatch := doc.Find(fmt.Sprintf("#%s", loginFormHtmlId)).First()
	return loginFormMatch.Length() == 0
}
