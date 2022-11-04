package parse

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"io"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

const ExamTitleUnknown = "Unknown Exam"

// ExaminationSchedule attempts to parse a page into a models.ExaminationSchedule model.
// This function expects the Amizone "Examination Schedule" page, parsable into an HTML document.
func ExaminationSchedule(body io.Reader) (*models.ExaminationSchedule, error) {
	const (
		breadcrumbsSelector    = "#breadcrumbs > ul.breadcrumb > li.active"
		scheduleBreadcrumbText = "Examination Schedule"
	)

	// "data-title" attributes for exams table entry cells
	const (
		dataCellSelectorTpl = "td[data-title='%s']"

		dTitleCode = "Course Code"
		dTitleName = "Course Title"
		dTitleDate = "Exam Date"
		dTitleTime = "Time"
		dTitleType = "Paper Type"
	)

	const (
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
	// @todo: Need tests with valid page that doesn't have exams information.
	scheduleTable := dom.Find("table.table")
	if scheduleTable.Length() == 0 {
		klog.Warning("Failed to find the examination exams table. What's up?")
		return nil, errors.New(ErrFailedToParse)
	}

	// Attempt to get the examination exams rows.
	scheduleEntries := scheduleTable.Find("tbody > tr")
	exams := make([]models.ScheduledExam, scheduleEntries.Length())

	// Iterate through exams rows to parse entries
	scheduleEntries.Each(func(i int, row *goquery.Selection) {
		exam := models.ScheduledExam{
			Course: models.CourseRef{
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

		exams[i] = exam
	})

	// Attempt to get the examination title.
	title := func() string {
		raw := dom.Find("div.page-header h1").Text()
		if raw != "" {
			sanitised := strings.TrimSpace(raw)
			// The title is usually like "EXAM TITLE ALL CAPS"
			title := strings.Title(strings.ToLower(sanitised))
			return title
		}
		klog.Warning("Failed to find the exam title. What's up?")
		return ExamTitleUnknown
	}()

	return &models.ExaminationSchedule{
		Title: title,
		Exams: exams,
	}, nil
}
