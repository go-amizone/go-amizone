package handlers

import (
	"amizone/amizone"
	"amizone/amizone_api/response_models"
	"net/http"
	"time"
)

func (a *handlerCfg) authenticatedAttendanceHandler(rw http.ResponseWriter, r *http.Request, c amizone.ClientInterface) {
	if r.Method == "GET" {
		attendance, err := c.GetAttendance()
		if err != nil {
			a.l.Error(err, "Failed to get attendance from the amizone client", "client", c)
			rw.WriteHeader(http.StatusInternalServerError)
		}
		err = WriteJsonResponse(attendance, rw)
		if err != nil {
			a.l.Error(err, "Failed to write attendance to the response writer", "client", c)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func (a *handlerCfg) authenticatedClassScheduleHandler(rw http.ResponseWriter, r *http.Request, c amizone.ClientInterface) {
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

func (a *handlerCfg) AttendanceHandler(rw http.ResponseWriter, r *http.Request) {
	authenticatedHandlerWrapper(a, a.authenticatedAttendanceHandler)(rw, r)
}

func (a *handlerCfg) ClassScheduleHandler(rw http.ResponseWriter, r *http.Request) {
	authenticatedHandlerWrapper(a, a.authenticatedClassScheduleHandler)(rw, r)
}
