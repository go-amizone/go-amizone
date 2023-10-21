package models

import "time"

// AtpcPlacementDetails is a model for representing ATPC placement details.
type AtpcPlacementEntry struct {
	Company     string
	RegStartDate time.Time
	RegEndDate   time.Time
}

type AtpcPlacementDetails []AtpcPlacementEntry