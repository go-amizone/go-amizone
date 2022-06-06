package parse

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"io"
	"k8s.io/klog/v2"
	"regexp"
	"strconv"
)

// Courses parses the Amizone courses page.
func Courses(body io.Reader) (models.Courses, error) {
	// selectors
	const (
		selectorPrimaryCourseTable   = "#CourseListSemWise > div:nth-child(1) > table:nth-child(1)"
		selectorSecondaryCourseTable = "#CourseListSemWise > div:nth-child(2) > table:nth-child(1)"
	)

	// "data-title" attributes for the primary course table
	const (
		dtCode        = "Course Code"
		dtName        = "Course Name"
		dtType        = "Type"
		dtSyllabusDoc = "Course Syllabus"
		dtAttendance  = "Attendance"
		dtInternals   = "Internal Asses."
	)

	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToParseDOM, err.Error()))
	}

	if !isCoursesPage(dom) {
		return nil, errors.New(ErrFailedToParse)
	}

	courseTablePrimary := dom.Find(selectorPrimaryCourseTable)
	if matches := courseTablePrimary.Length(); matches != 1 {
		klog.Warning("failed to find the main course table. selector matches:", matches)
		return nil, errors.New(ErrFailedToParse)
	}

	// primary courses
	primaryEntries := courseTablePrimary.Find(selectorDataRows)
	if primaryEntries.Length() == 0 {
		klog.Errorf("found no primary courses on the courses page")
		return nil, errors.New(ErrFailedToParse)
	}

	// secondary courses
	secondaryEntries := dom.Find(selectorSecondaryCourseTable).Find(selectorDataRows)

	// all courses
	courseEntries := primaryEntries.AddSelection(secondaryEntries)

	// Build up our entries
	courses := make(models.Courses, courseEntries.Length())
	courseEntries.Each(func(i int, row *goquery.Selection) {
		course := models.Course{
			CourseRef: models.CourseRef{
				Name: cleanString(row.Find(fmt.Sprintf(selectorTplDataCell, dtName)).Text()),
				Code: cleanString(row.Find(fmt.Sprintf(selectorTplDataCell, dtCode)).Text()),
			},
			Type: cleanString(row.Find(fmt.Sprintf(selectorTplDataCell, dtType)).Text()),
			Attendance: func() models.Attendance {
				raw := row.Find(fmt.Sprintf(selectorTplDataCell, dtAttendance)).Text()
				// go std regex doesn't have lookarounds :(
				attendedStr := regexp.MustCompile(`\d{1,2}/`).FindString(raw)
				attended, err1 := strconv.Atoi(cleanString(attendedStr, '/'))
				totalStr := regexp.MustCompile(`/\d{1,2}`).FindString(raw)
				total, err2 := strconv.Atoi(cleanString(totalStr, '/'))
				if err1 != nil || err2 != nil {
					klog.Warning("parse(courses): attendance string has unexpected format")
					return models.Attendance{}
				}
				return models.Attendance{
					ClassesAttended: attended,
					ClassesHeld:     total,
				}
			}(),
			InternalMarks: func() models.Marks {
				raw := row.Find(fmt.Sprintf(selectorTplDataCell, dtInternals)).Text()
				gotStr := regexp.MustCompile(`\d{1,2}(\.\d{1,2})?[\[/]`).FindString(raw)
				got, err1 := strconv.ParseFloat(cleanString(gotStr, '[', '/'), 32)
				maxStr := regexp.MustCompile(`/\d{1,2}(\.\d{1,2})?`).FindString(raw)
				max, err2 := strconv.ParseFloat(cleanString(maxStr, '/'), 32)
				// @todo make allowances if marks aren't there!??
				if err1 != nil || err2 != nil {
					klog.Warning("parse(courses): error in parsing marks")
					return models.Marks{}
				}
				return models.Marks{
					Max:  float32(max),
					Have: float32(got),
				}
			}(),
			SyllabusDoc: row.Find(fmt.Sprintf(selectorTplDataCell, dtSyllabusDoc)).Find("a").AttrOr("href", ""),
		}
		courses[i] = course
	})

	return courses, nil
}

func isCoursesPage(dom *goquery.Document) bool {
	const coursePageBreadcrumb = "My Courses"
	return dom.Find(selectorActiveBreadcrumb).Text() == coursePageBreadcrumb
}
