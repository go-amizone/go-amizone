package parse

import "strings"

// Errors
const (
	ErrFailedToParse    = "failed to parse"
	ErrFailedToParseDOM = ErrFailedToParse + " DOM"
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
	return strings.Trim(ws, string(set))
}
