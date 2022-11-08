# Contributing to go-amizone

## Reporting Bugs

If you find a bug, you can contribute by creating an issue on [GitHub](https://github.com/ditsuke/go-amizone).
After reporting you can choose to work on the issue yourself or wait for another contributor to pick it up.
If you choose to pick it up yourself, please leave a comment to let others know. We're also happy to help you
get started if you need any help!

## Submitting a new feature or fix

If you find an issue that you would like to fix or a new feature you want to work on, please open a new GitHub
[Pull Request](https://github.com/ditsuke/go-amizone/pulls) with your changes. Please attempt to follow the same
coding style as the rest of the project and make sure to include tests for your changes and ensure all tests pass.

## Development

### Prerequisites

- [Go](https://golang.org/dl/) 1.18 or later.
- [Buf](https://docs.buf.build/installation) 1.9 or later. Buf is a tool for linting and generating code from protobuf files,
  so you'll only need it if you want to make changes to the backend API (adding a new endpoint, for example).

On Windows, you will need to install [Git Bash](https://gitforwindows.org/) to run the make commands.

### The File Structure

The project is structured as follows:

```text
├── amizone
│   └── internal
│       ├── mock
│       ├── models
│       └── parse
├── cmd
│   └── amizone-api-server
├── hack
│   └── tools
└── server
    ├── gen
    │   ├── go
    │   └── openapiv2
    ├── proto
    │   └── v1
    └── transformers
        ├── fromproto
        └── toproto

```

Largely, the project is divided into two parts: the Amizone SDK (`amizone`) and the API server (`server`).

- The SDK is a Go package that provides a simple interface for interacting with Amizone. It is located in the `amizone` directory.
- The API server is a multiplexed server that uses the SDK and exposes the API as a gRPC service and a REST API. It uses
  protobufs to define data structures, gRPC services and REST endpoints, which are then used to generate Go stubs and OpenAPI
  definitions using the [Buf](https://buf.build) tool.

The `cmd` directory contains the code for the `amizone-api-server` binary, which is the entrypoint for the API server.

Finally, the `hack` directory contains some setup code to install dependencies for the project (mainly protoc plugins for the Buf tool).

### Making changes to the API Server

If you're making changes to the API server, you'll likely make changes to the following files:

- `server/proto/v1/amizone.proto`: Contains definitions for API endpoints.
- `server/grpc_service_server.go`: Contains bridging code between the gRPC service and the SDK.

> **Note**
>
> After making changes to the protobuf files, you'll need to run `make generate-proto` to regenerate the Go stubs and OpenAPI definitions.

### Running tests

To run the tests locally, run `make test-unit` for unit tests and `make test-integration` for integration tests.

> **Note**
>
> Integration tests require a valid set of Amizone credentials to run. You can set the credentials in the `.env` file by copying the `.env.sample` file and filling in your credentials.
