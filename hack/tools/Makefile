SHELL := /usr/bin/env bash
GO ?= go
GOPATH := $(shell $(GO) env GOPATH)

PROTOC ?= protoc
GOGOPROTO_VERSION ?= v1.3.2
GOGOPROTOPATH ?= $(GOPATH)/pkg/mod/github.com/gogo/protobuf@$(GOGOPROTO_VERSION)

.PHONY: tools
tools:
	@echo "Installing tools..."
	${GO} install \
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway" \
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2" \
	"google.golang.org/grpc/cmd/protoc-gen-go-grpc" \
	"google.golang.org/protobuf/cmd/protoc-gen-go" \
