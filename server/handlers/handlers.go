package handlers

import (
	"github.com/ditsuke/go-amizone/amizone"
	"github.com/go-logr/logr"
	"net/http"
)

// Cfg is a handler configuration struct for the API.
// It currently provides the handlers access to the logger configured by the main application.
type Cfg struct {
	L logr.Logger
	A amizone.ClientFactoryInterface
}

func NewCfg(l logr.Logger, a amizone.ClientFactoryInterface) *Cfg {
	return &Cfg{L: l, A: a}
}

func (a *Cfg) AttendanceHandler(rw http.ResponseWriter, r *http.Request) {
	authenticatedHandlerWrapper(a, a.authenticatedAttendanceHandler)(rw, r)
}

func (a *Cfg) ClassScheduleHandler(rw http.ResponseWriter, r *http.Request) {
	authenticatedHandlerWrapper(a, a.authenticatedClassScheduleHandler)(rw, r)
}

func (a *Cfg) ExamScheduleHandler(rw http.ResponseWriter, r *http.Request) {
	authenticatedHandlerWrapper(a, a.authenticatedExamScheduleHandler)(rw, r)
}
