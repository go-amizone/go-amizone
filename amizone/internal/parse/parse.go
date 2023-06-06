package parse

import (
	"html"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

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

// UnescapeUnicode unescapes unicode characters in a string.
// Ref: https://groups.google.com/g/golang-nuts/c/KO1yubIbKpU/m/ue_EU8dcBQAJ
func UnescapeUnicode(s string) string {
	quoted := strconv.Quote(s)
	unicodeRepl := strings.ReplaceAll(quoted, `\\u`, `\u`)
	unquoted, err := strconv.Unquote(unicodeRepl)
	if err != nil {
		return err.Error()
	}
	return unquoted
}

// CleanString trims off whitespace and additional runes passed.
func CleanString(s string, set ...rune) string {
	p := bluemonday.NewPolicy()
	// amizone (sometimes) sends certain some utf8 characters encoded
	unicode := UnescapeUnicode(s)
	// amizone sometimes sends markup mixed with strings
	htmlSanitized := html.UnescapeString(p.Sanitize(html.UnescapeString(unicode)))
	ws := strings.TrimSpace(htmlSanitized)
	wd := strings.Trim(ws, string(set))
	return strings.TrimSpace(wd)
}
