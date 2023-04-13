package parse

import (
	"errors"
	"fmt"
	"io"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/models"
)

func isFacultyPage(dom *goquery.Document) bool {
	const FacultyPageBreadcrumb = "My Faculty"
	return CleanString(dom.Find(selectorActiveBreadcrumb).Text()) == FacultyPageBreadcrumb
}

func FacultyFeedback(body io.Reader) (models.FacultyFeedbackSpecs, error) {
	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrFailedToParseDOM, err)
	}

	if !IsLoggedInDOM(dom) {
		return nil, errors.New(ErrNotLoggedIn)
	}

	if !isFacultyPage(dom) {
		return nil, fmt.Errorf("%s: Not Faculty Feedback Page", ErrFailedToParse)
	}

	specs := make(models.FacultyFeedbackSpecs, 0)
	dom.Find("i[title='Please click here to give faculty feedback']").Each(func(_ int, opt *goquery.Selection) {
		parentAnchor := opt.Parent()
		if parentAnchor.Length() == 0 {
			// log
			return
		}
		uri, err := url.Parse(parentAnchor.AttrOr("href", ""))
		if err != nil {
			// log
			return
		}
		specs = append(specs, models.FacultyFeedbackSpec{
			VerificationToken: VerificationTokenFromDom(dom),
			FacultyId:         url.QueryEscape(uri.Query().Get("FacultyStaffID")),
			CourseType:        url.QueryEscape(uri.Query().Get("CourseType")),
			DepartmentId:      url.QueryEscape(uri.Query().Get("DetID")),
			SerialNumber:      url.QueryEscape(uri.Query().Get("SrNo")),
		})
	})

	return specs, nil
}
