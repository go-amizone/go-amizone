package amizone

import (
	"net/http"
)

// ClientInterface is an exported interface for amizoneClient to make mocking and testing more convenient.
type ClientInterface interface {
	DidLogin() bool
	GetAttendance() (Attendance, error)
	GetClassSchedule(date Date) (ClassSchedule, error)
	GetExamSchedule() (*ExamSchedule, error)
}

// Interface compliance constraint for amizoneClient
var _ ClientInterface = &amizoneClient{}

// ClientFactoryInterface is a type for functions that return ClientInterface
// instances. Functions returning concrete types need to be wrapped by functions
// that return the interface; apparently a limitation of the Go compiler's type
// inference.
type ClientFactoryInterface func(cred Credentials, httpClient *http.Client) (ClientInterface, error)

// Interface compliance constraint for NewClient [requires wrapping]
var _ ClientFactoryInterface = func(cred Credentials, httpClient *http.Client) (ClientInterface, error) {
	return NewClient(cred, httpClient)
}
