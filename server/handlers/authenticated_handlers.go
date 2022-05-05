package handlers

import (
	"github.com/ditsuke/go-amizone/amizone"
	"github.com/ditsuke/go-amizone/server/response_models"
	"net/http"
	"time"
)

// AuthenticatedHandler functions handle requests to Amizone that require auth.
// They need to be wrapped by a decorator that checks the auth parameters and creates an amizone.ClientInterface
// instance before calling onto the authenticated handler.
type AuthenticatedHandler func(
	rw http.ResponseWriter,
	r *http.Request,
	c amizone.ClientInterface)

func (a *Cfg) authenticatedAttendanceHandler(rw http.ResponseWriter, r *http.Request, c amizone.ClientInterface) {
	if r.Method == "GET" {
		attendance, err := c.GetAttendance()
		if err != nil {
			a.L.Error(err, "Failed to get attendance from the amizone client", "client", c)
			rw.WriteHeader(http.StatusInternalServerError)
		}
		err = WriteJsonResponse(attendance, rw)
		if err != nil {
			a.L.Error(err, "Failed to write attendance to the response writer", "client", c)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (a *Cfg) authenticatedClassScheduleHandler(rw http.ResponseWriter, r *http.Request, c amizone.ClientInterface) {
	var err error
	if r.Method == "GET" {
		t := time.Now()
		if date := r.FormValue("date"); date != "" {
			t, err = time.Parse("2006-01-02", date)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				_ = WriteJsonResponse(response_models.ErrorResponse{{Message: "Invalid date format"}}, rw)
				return
			}
		}
		schedule, err := c.GetClassSchedule(amizone.DateFromTime(t))
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		_ = WriteJsonResponse(schedule, rw)
	}
}

func (a *Cfg) authenticatedExamScheduleHandler(rw http.ResponseWriter, r *http.Request, c amizone.ClientInterface) {
	if r.Method == http.MethodGet {
		schedule, err := c.GetExamSchedule()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
		}
		err = WriteJsonResponse(schedule, rw)
		if err != nil {
			a.L.Error(err, "Failed to write exam schedule to the response writer", "client", c)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}
