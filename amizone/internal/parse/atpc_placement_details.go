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

func isAtpcPlacementDetailsPage(dom *goquery.Document) bool {
	PlacementDetailsPageBreadcrumb := "Placement Details"
	return CleanString(dom.Find(selectorActiveBreadcrumb).Filter(fmt.Sprintf(":contains('%s')", PlacementDetailsPageBreadcrumb)).Text()) == PlacementDetailsPageBreadcrumb
}

func AtpcPlacementDetails(body io.Reader) (models.AtpcPlacementDetails, error) {
	// selectors
	const (
		selectorDataRows = "tbody > tr"
	)

	dom, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", ErrFailedToParseDOM, err)
	}

	if !IsLoggedInDOM(dom) {
		return nil, errors.New(ErrNotLoggedIn)
	}

	if !isAtpcPlacementDetailsPage(dom) {
		klog.Warning("failed to find the placement details page breadcrumb")
		return nil, errors.New(ErrFailedToParse)
	}

	placementDetailsTable := dom.Find("table.table")
	if placementDetailsTable.Length() == 0 {
		klog.Warning("failed to find the placement details table")
		return nil, errors.New(ErrFailedToParse)
	}

	// placement details
	placementDetailsEntries := placementDetailsTable.Find(selectorDataRows)
	if placementDetailsEntries.Length() == 0 {
		klog.Errorf("found no placement details")
		return nil, errors.New(ErrFailedToParse)
	}

	// Build up our entries
	placementDetails := make(models.AtpcPlacementDetails, placementDetailsEntries.Length())
	placementDetailsEntries.Each(func(i int, row *goquery.Selection) {
		placementDetail := models.AtpcPlacementEntry{
			Company: CleanString(row.Find("td:nth-child(2)").Text()),
			RegStartDate: func() time.Time {
				date, err := time.Parse("02/01/2006", CleanString(row.Find("td:nth-child(3)").Text()))
				if err != nil {
					klog.Errorf("failed to parse placement details' registration start date: %s", err)
				}
				return date
			}(),
			RegEndDate: func() time.Time {
				date, err := time.Parse("02/01/2006", CleanString(row.Find("td:nth-child(4)").Text()))
				if err != nil {
					klog.Errorf("failed to parse placement details' registration end date: %s", err)
				}
				return date
			}(),
		}

		placementDetails[i] = placementDetail
	})

	return placementDetails, nil
}
