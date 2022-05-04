package handlers

import (
	"github.com/go-logr/logr"
)

// handlerCfg is a handler configuration struct for the API.
// It currently provides the handlers access to the logger configured by the main application.
type handlerCfg struct {
	l logr.Logger
}

func NewHandlerCfg(l logr.Logger) *handlerCfg {
	return &handlerCfg{l: l}
}
