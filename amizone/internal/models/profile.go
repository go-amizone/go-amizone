package models

import "time"

// Profile models information exposed by the Amizone ID card page
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
