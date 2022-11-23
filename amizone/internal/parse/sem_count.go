package parse

import (
	"errors"
	"io"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
)

// Semesters returns the number of ongoing or passed semesters from the Amizone courses page.
func Semesters(body io.Reader) (models.SemesterList, error) {
	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, errors.New(ErrFailedToParseDOM)
	}

	if !isCoursesPage(dom) {
		return nil, errors.New(ErrFailedToParse)
	}

	var semesters models.SemesterList
	dom.Find("#CurrentSemesterInfo option").Each(func(_ int, opt *goquery.Selection) {
		if value := opt.AttrOr("value", ""); value != "" {
			sem := models.Semester{
				Name: cleanString(opt.Text()),
				Ref:  value,
			}
			semesters = append(semesters, sem)
		}
	})

	return semesters, nil
}
