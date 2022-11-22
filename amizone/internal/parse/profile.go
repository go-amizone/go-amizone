package parse

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/internal/models"
	"k8s.io/klog/v2"
)

func Profile(body io.Reader) (*models.Profile, error) {
	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", ErrFailedToParseDOM, err)
	}

	if !isIDCardPage(dom) {
		return nil, errors.New(ErrFailedToParse)
	}

	const (
		selectorCardFront = "#lblNameIDCardFront1"
		selectorCardBack  = "#lblInfoIDCardBack1"
		selectorHeadshot  = "img#ImgPhotoIDCardFront1"
	)

	name, course, batch := func() (string, string, string) {
		conDiv := dom.Find(selectorCardFront)
		// Replace <br>'s with newlines to make the semantic soup parsable
		conDiv.Find("br").ReplaceWithHtml("\n")
		all := cleanString(conDiv.Text())
		allSlice := strings.Split(all, "\n")
		if len(allSlice) != 3 {
			klog.Error("failed to parse out name, course and batch from the ID page")
			return "", "", ""
		}

		for i, s := range allSlice {
			allSlice[i] = cleanString(s)
		}

		return allSlice[0], allSlice[1], allSlice[2]
	}()

	// We now have some basic information to populate
	profile := &models.Profile{
		Name:    name,
		Program: course,
		Batch:   batch,
	}

	// Parse "SUID": a student UUID
	profile.UUID = func() string {
		headshotUrl, exists := dom.Find(selectorHeadshot).Attr("src")
		if !exists {
			klog.Warning("parse(profile): could not find profile student headshot URL")
			return ""
		}
		studentUUID := regexp.MustCompile(`\w{8}-\w{4}-\w{4}-\w{4}-\w{12}`).FindString(headshotUrl)
		if studentUUID == "" {
			klog.Warning("parse(profile): could not find student uuid in headshot URL")
		}
		return studentUUID
	}()

	const (
		lblEnrollmentNo = "Enrollment No"
		lblDOB          = "Date Of Birth"
		lblBloodGroup   = "Blood Group"
		lblValidity     = "Validity"
		lblCardNo       = "ID Card No"

		timeFormat = "02.01.2006"
	)

	// Parse stuff from "back" of the card
	backDiv := dom.Find(selectorCardBack)
	// replace <br>'s with newlines
	backDiv.Find("br").ReplaceWithHtml("\n")
	everything := strings.Split(
		cleanString(backDiv.Text()),
		"\n",
	)

	labelRegexp := regexp.MustCompile(`[\w .]+( )?:`)
	valueRegexp := regexp.MustCompile(`:( )?.*$`)

	for _, line := range everything {
		lbl := cleanString(labelRegexp.FindString(line), ':')
		value := cleanString(valueRegexp.FindString(line), ':')
		switch lbl {
		case lblEnrollmentNo:
			profile.EnrollmentNumber = value
		case lblDOB:
			dob, err := time.Parse(timeFormat, value)
			if err != nil {
				klog.Warningf("failed to parse DOB from ID card: %v", err)
				break
			}
			profile.DateOfBirth = dob
		case lblBloodGroup:
			profile.BloodGroup = value
		case lblValidity:
			validity, err := time.Parse(timeFormat, value)
			if err != nil {
				klog.Warningf("failed to parse validity from ID card: %v", err)
				break
			}
			profile.EnrollmentValidity = validity
		case lblCardNo:
			profile.IDCardNumber = value
		}
	}

	return profile, nil
}

func isIDCardPage(dom *goquery.Document) bool {
	const IDCardPageBreadcrumb = "ID Card View"
	return cleanString(dom.Find(selectorActiveBreadcrumb).Text()) == IDCardPageBreadcrumb
}
