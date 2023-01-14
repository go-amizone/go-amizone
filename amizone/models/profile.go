package models

import "time"

// Profile is a model for representing a user's Amizone profile.
type Profile struct {
	Name               string
	EnrollmentNumber   string
	EnrollmentValidity time.Time
	Batch              string
	Program            string
	DateOfBirth        time.Time
	BloodGroup         string
	IDCardNumber       string
	UUID               string
}
