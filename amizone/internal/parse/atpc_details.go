package parse

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ditsuke/go-amizone/amizone/models"
	"k8s.io/klog/v2"
)

func isAtpcPage(dom *goquery.Document) bool {
	PlacementDetailsPageBreadcrumb := "Placement Details"
	CorporateDetailsPageBreadcrumb := "Corporate Details"
	activeBreadcrumb := dom.Find(selectorActiveBreadcrumb)
	isPlacementDetailsPage := CleanString(activeBreadcrumb.Filter(fmt.Sprintf(":contains('%s')", PlacementDetailsPageBreadcrumb)).Text()) == PlacementDetailsPageBreadcrumb
	isCorporateDetailsPage := CleanString(activeBreadcrumb.Filter(fmt.Sprintf(":contains('%s')", CorporateDetailsPageBreadcrumb)).Text()) == CorporateDetailsPageBreadcrumb
	return isPlacementDetailsPage || isCorporateDetailsPage
}

func AtpcDetails(body io.Reader) (models.AtpcDetails, error) {
	// selectors
	const (
		selectorDataRows = "tbody > tr"
		selectorNotFoundText     = "div.card div.card-body > form > h1" // used for searching "not applicable" text, in case of no atpc details table
	)

	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToParseDOM, err)
	}

	if !IsLoggedInDOM(dom) {
		return nil, errors.New(ErrNotLoggedIn)
	}

	if !isAtpcPage(dom) {
		klog.Warning("failed to find the atpc details page breadcrumb")
		return nil, errors.New(ErrFailedToParse)
	}

	atpcDetailsTable := dom.Find("table.table")
	if atpcDetailsTable.Length() == 0 {
		atpcNotFoundText := CleanString(dom.Find(selectorNotFoundText).Text())
		if atpcNotFoundText == "Not Applicable" {
			return nil, nil
		} else {
			klog.Warning("failed to find the atpc details table")
			return nil, errors.New(ErrFailedToParse)
		}
	}

	// atpc details
	atpcDetailsEntries := atpcDetailsTable.Find(selectorDataRows)
	if atpcDetailsEntries.Length() == 0 {
		klog.Errorf("found no atpc details")
		return nil, errors.New(ErrFailedToParse)
	}

	// Build up our entries
	atpcDetails := make(models.AtpcDetails, atpcDetailsEntries.Length())
	atpcDetailsEntries.Each(func(i int, row *goquery.Selection) {
		atpcDetail := models.AtpcEntry{
			Company: CleanString(row.Find("td:nth-child(2)").Text()),
			RegStartDate: func() time.Time {
				date, err := time.Parse("02/01/2006", CleanString(row.Find("td:nth-child(3)").Text()))
				if err != nil {
					klog.Errorf("failed to parse atpc details' registration start date: %s", err)
				}
				return date
			}(),
			RegEndDate: func() time.Time {
				date, err := time.Parse("02/01/2006", CleanString(row.Find("td:nth-child(4)").Text()))
				if err != nil {
					klog.Errorf("failed to parse atpc details' registration end date: %s", err)
				}
				return date
			}(),
		}

		atpcDetails[i] = atpcDetail
	})

	return atpcDetails, nil
}
