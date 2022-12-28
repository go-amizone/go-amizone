package parse

import "strings"

// Errors
const (
	ErrFailedToParse    = "failed to parse"
	ErrFailedToParseDOM = ErrFailedToParse + " DOM"
	ErrNotLoggedIn      = ErrFailedToParse + ": not logged in"
)

// Shared selectors
const (
	selectorActiveBreadcrumb = "ul.breadcrumb li.active"
	selectorDataRows         = "tbody > tr"
)

// Shared selector templates
const (
	selectorTplDataCell = "td[data-title='%s']"
)

// Utilities

// cleanString trims off whitespace and additional runes passed.
func cleanString(s string, set ...rune) string {
	ws := strings.TrimSpace(s)
	wd := strings.Trim(ws, string(set))
	return strings.TrimSpace(wd)
}
