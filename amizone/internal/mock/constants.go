package mock

import "time"

type Timestamp int64

func (t Timestamp) Time() time.Time {
	return time.Unix(int64(t), 0).UTC()
}

// Constants for use in tests using the mock package to create Gock requests, etc.
const (
	ValidUser = "fakeUsername"
	ValidPass = "fakePassword"

	InvalidUser = "this-user-does-not-exist"
	InvalidPass = "this-password-does-not-either"

	AuthCookie        = "fakeAuthCookie"
	VerificationToken = "fakeRequestVerificationToken"
	SessionID         = "fakeSessionId"

	// StudentUUID is the UUID associated with the student used across testdata in filesystem.
	StudentUUID = "98RFGK88-A01C-1JJO-N73D-4BJR42B33J51"

	StudentName             = "John Doe"
	StudentEnrollmentNumber = "A2305221007"
	StudentIDCardNumber     = "95188911"
	StudentBloodGroup       = "B-ve"
	StudentProgram          = "B.Tech (CSE)"
	StudentBatch            = "2020-2024"

	StudentIDValidity Timestamp = 1719705600
	StudentDOB        Timestamp = 986428800
)
