SHELL := /usr/bin/env bash

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))

GO ?= go
GOPATH := $(shell $(GO) env GOPATH)

export GOBIN := $(dir ${MKFILE_PATH})bin

.PHONY: tools
tools:
	@echo "Installing tools..."
	${GO} install \
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway" \
	"github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2" \
	"google.golang.org/grpc/cmd/protoc-gen-go-grpc" \
	"google.golang.org/protobuf/cmd/protoc-gen-go" \
	"github.com/rakyll/gotest"
