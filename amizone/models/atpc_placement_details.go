package models

import "time"

// AtpcDetails is a model for representing ATPC details.
type AtpcEntry struct {
	Company      	string
	RegStartDate 	time.Time
	RegEndDate   	time.Time
}

type AtpcDetails []AtpcEntry

type AtpcListings struct {
	Placement 			AtpcDetails
	Internship 			AtpcDetails
	CorporateEvent 		AtpcDetails
}