package server

import (
	"context"
	"github.com/ditsuke/go-amizone/amizone"
	"github.com/ditsuke/go-amizone/server/handlers"
	"github.com/go-logr/logr"
	"net/http"
)

type AmizoneClientFactory func(amizone.Credentials) amizone.ClientInterface

type Config struct {
	Logger   logr.Logger
	BindAddr string
}

func NewConfig() *Config {
	return &Config{
		BindAddr: "127.0.0.1:8081",
	}
}

type ApiServer struct {
	Config     *Config
	Router     *http.ServeMux
	httpServer *http.Server
}

func New(config *Config) *ApiServer {
	return &ApiServer{
		Config: config,
		Router: http.NewServeMux(),
	}
}

func (s *ApiServer) Run() error {
	if s.httpServer == nil {
		s.Config.Logger.V(1).Info("Initializing server...")
		s.httpServer = &http.Server{
			Addr:    s.Config.BindAddr,
			Handler: s.Router,
		}
	}

	s.Config.Logger.Info("Starting server", "bind_addr", s.Config.BindAddr)
	s.configureRouter()
	return s.httpServer.ListenAndServe()
}

func (s *ApiServer) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *ApiServer) configureRouter() {
	handlerCfg := handlers.Cfg{
		L: s.Config.Logger,
		A: func(cred amizone.Credentials, httpClient *http.Client) (amizone.ClientInterface, error) {
			return amizone.NewClient(cred, httpClient)
		}}

	s.Router.HandleFunc("/attendance", handlerCfg.AttendanceHandler)
	s.Router.HandleFunc("/class_schedule", handlerCfg.ClassScheduleHandler)
	s.Router.HandleFunc("/exam_schedule", handlerCfg.ExamScheduleHandler)
}
