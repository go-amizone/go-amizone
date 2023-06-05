package parse

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/models"
	"k8s.io/klog/v2"
)

// ExaminationResult attempts to parse exam result information from the Amizone Examination Results page
// into a models.ExaminationResultRecords instance.
func ExaminationResult(body io.Reader) (*models.ExamResultRecords, error) {
	const (
		resultTablesSelector = "div#no-more-tables"
		coursesResultIndex   = 0
		overallResultIndex   = 1
	)

	// "data-title" attributes for exams result entry cells
	const (
		dataCellSelectorTpl = "td[data-title='%s']"

		dTitleCode = "Course Code"
		dTitleName = "Course Title"

		dTitleMax  = "Max Total"
		dTitleAcu  = "ACU"
		dTitleGo   = "Go"
		dTitleGp   = "GP"
		dTitleCp   = "CP"
		dTitleEcu  = "ECU"
		dTitleDate = "PublishDate"

		dTitleSem  = "Semester"
		dTitleSgpa = "SGPA"
		dTitleCgpa = "CGPA"
		dTitleBack = "Back Papers"
	)

	const (
		// format for time.Parse() after appending date and time from the table
		tableDateFormat = "02/01/2006"
	)

	dom, err := goquery.NewDocumentFromReader(body)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToParseDOM, err)
	}

	if !IsLoggedInDOM(dom) {
		return nil, errors.New(ErrNotLoggedIn)
	}

	// Try to find the two tables to see if we are on the correct page
	tables := dom.Find(resultTablesSelector).Children()
	if tables.Length() != 2 {
		klog.Warning("Wrong number of tables detected in 'Examination Result'. Are we on the right page and logged in?")
		return nil, errors.New(ErrFailedToParse)
	}

	// Get the table body from the <div>
	courseWiseResultTable := tables.Eq(coursesResultIndex).Find("tbody")
	overallResultTable := tables.Eq(overallResultIndex).Find("tbody")

	// Gets every <tr> from the table
	overallResultEntries := overallResultTable.Children()
	overallResult := make([]models.OverallResult, overallResultEntries.Length())
	overallResultEntries.Each(func(i int, row *goquery.Selection) {
		result := models.OverallResult{
			Semester: models.Semester{
				Name: row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleSem)).Text(),
				Ref:  row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleSem)).Text(),
			},
			SemesterGradePointAverage:      float32(parseToFloat(CleanString(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleSgpa)).Text()))),
			CummulatitiveGradePointAverage: float32(parseToFloat((CleanString(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleCgpa)).Text())))),
		}
		overallResult[i] = result
	})

	// Gets every <tr> from the table
	courseWiseResultEntries := courseWiseResultTable.Children()
	courseWiseResult := make([]models.ExamResultRecord, courseWiseResultEntries.Length())
	courseWiseResultEntries.Each(func(i int, row *goquery.Selection) {
		result := models.ExamResultRecord{
			Course: models.CourseRef{
				Code: row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleCode)).Text(),
				Name: row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleName)).Text(),
			},
			Result: models.CourseResult{
				MaxTotal:             parseToInt(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleMax)).Text()),
				AquiredCreditUnits:   parseToInt(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleAcu)).Text()),
				GradeObtained:        row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleGo)).Text(),
				GradePoint:           parseToInt(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleGp)).Text()),
				CreditPoints:         parseToInt(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleCp)).Text()),
				EffectiveCreditUnits: parseToInt(row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleEcu)).Text()),
				PublishDate: func() time.Time {
					parsedTime, nil := time.Parse(tableDateFormat, row.Find(fmt.Sprintf(dataCellSelectorTpl, dTitleDate)).Text())
					if err != nil {
						klog.Warningf("Failed to parse publish date: %s", err.Error())
					}
					return parsedTime
				}(),
			},
		}
		courseWiseResult[i] = result
	})

	resultRecords := models.ExamResultRecords{
		CourseWise: courseWiseResult,
		Overall:    overallResult,
	}

	return &resultRecords, nil
}

// parseToFloat parses an integer to a string, logs on failure.
func parseToFloat(raw string) float64 {
	if raw == "" {
		return 0.0
	}
	i, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		klog.Errorf("Failed to parse string to float: %s", err.Error())
	}
	return i
}
