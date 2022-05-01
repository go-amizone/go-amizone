package parse

import (
	"GoFriday/internal/models"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

// ExaminationSchedule attempts to parse a page into a models.ExaminationSchedule model.
// This function expects the Amizone "Examination Schedule" page, parsable into an HTML document.
func ExaminationSchedule(body io.Reader) (models.ExaminationSchedule, error) {
	const (
		breadcrumbsSelector    = "#breadcrumbs > ul.breadcrumb > li.active"
		scheduleBreadcrumbText = "Examination Schedule"
	)

	// "data-title" attributes for schedule table entry cells
	const (
		dTitleCode = "Course Code"
		dTitleName = "Course Title"
		dTitleDate = "Exam Date"
		dTitleTime = "Time"
		dTitleType = "Paper Type"
	)

	const (
		dataCellSelectorTpl = "td[data-title='%s']"
		// format for time.Parse() after appending date and time from the table
		tableTimeFormat = "02/01/2006 15:04"
	)

	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToParseDOM, err.Error()))
	}

	// Try to find the "Examination Schedule" breadcrumb to determine if we're on the right page.
	if scheduleBreadcrumb := dom.Find(breadcrumbsSelector).
		Filter(fmt.Sprintf(":contains('%s')", scheduleBreadcrumbText)); scheduleBreadcrumb.Length() == 0 {
		klog.Warning("Failed to find the 'Examination Schedule' breadcrumb. Are we on the right page and logged in?")
		return nil, errors.New(ErrFailedToParse)
	}

	// Attempt to get the examination table.
	// @todo: Need tests with valid page that doesn't have schedule information.
	scheduleTable := dom.Find("table.table")
	if scheduleTable.Length() == 0 {
		klog.Warning("Failed to find the examination schedule table. What's up?")
		return nil, errors.New(ErrFailedToParse)
	}

	// Attempt to get the examination schedule rows.
	scheduleEntries := scheduleTable.Find("tbody > tr")
	schedule := make(models.ExaminationSchedule, scheduleEntries.Length())

	// Iterate through schedule rows to parse entries
	scheduleEntries.Each(func(i int, row *goquery.Selection) {
		exam := models.ScheduledExam{
			Course: &models.Course{
				Code: row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleCode)).Text(),
				Name: row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleName)).Text(),
			},
			Time: func() time.Time {
				rawDate := row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleDate)).Text()
				rawTime := row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleTime)).Text()
				parsedTime, err := time.Parse(tableTimeFormat, fmt.Sprintf("%s %s", rawDate, rawTime))
				if err != nil {
					klog.Warningf("Failed to parse exam time: %s", err.Error())
				}
				return parsedTime
			}(),
			Mode: func() string {
				raw := row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleType)).Text()
				return strings.TrimSpace(raw)
			}(),
		}

		schedule[i] = &exam
	})

	return schedule, nil
}
