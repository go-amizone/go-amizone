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

// Expose these data-title attributes, because they're used by the isCoursesPage function.
const (
	dtCourseCode       = "Course Code"
	dtCourseAttendance = "Attendance"
)

// Courses parses the Amizone courses page.
func Courses(body io.Reader) (models.Courses, error) {
	// selectors
	const (
		selectorPrimaryCourseTable   = "div:nth-child(1) > table:nth-child(1)"
		selectorSecondaryCourseTable = "div:nth-child(2) > table:nth-child(1)"
	)

	// "data-title" attributes for the primary course table
	const (
		dtCode        = dtCourseCode
		dtName        = "Course Name"
		dtType        = "Type"
		dtSyllabusDoc = "Course Syllabus"
		dtAttendance  = dtCourseAttendance
		dtInternals   = "Internal Asses."
	)

	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%s: %s", ErrFailedToParseDOM, err.Error()))
	}

	// We check for the course page first, but we can't rely on it alone because the "semester wise" course page does
	// not come with breadcrumbs.
	if !isCoursesPage(dom) {
		return nil, errors.New(ErrFailedToParse)
	}

	normDom := normalisePage(dom.Selection)

	courseTablePrimary := normDom.Find(selectorPrimaryCourseTable)
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
	secondaryEntries := normDom.Find(selectorSecondaryCourseTable).Find(selectorDataRows)

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
					ClassesAttended: int32(attended),
					ClassesHeld:     int32(total),
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

	return dom.Find(selectorActiveBreadcrumb).Text() == coursePageBreadcrumb ||
		(dom.Find(fmt.Sprintf(selectorTplDataCell, dtCourseCode)).Length() != 0 &&
			dom.Find(fmt.Sprintf(selectorTplDataCell, dtCourseAttendance)).Length() != 0)
}

// normalisePage attempts to "normalise" the page by extracting the contexts of the "#CourseListSemWise" div.
// We need to do this because the page comes in two flavors: one when it has breadcrumbs and the course tables wrapped
// in the "#CourseListSemWise" div, and one when it doesn't (when we query courses for a non-current semester).
func normalisePage(dom *goquery.Selection) *goquery.Selection {
	if child := dom.Find("#CourseListSemWise").Children(); child.Length() > 0 {
		return child
	}
	return dom
}
