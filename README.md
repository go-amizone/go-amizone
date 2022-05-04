# go-amizone
![Unit Tests](https://img.shields.io/github/workflow/status/ditsuke/go-amizone/test?label=tests&logo=github)
[![Coverage Status](https://img.shields.io/coveralls/github/ditsuke/go-amizone?logo=coveralls)](https://coveralls.io/github/ditsuke/go-amizone?branch=main)
![Issues](https://img.shields.io/github/issues/ditsuke/go-amizone?logo=github)
![License](https://img.shields.io/github/license/ditsuke/go-amizone)

**go-amizone** is a simple Go library and API client for the [Amizone](https://s.amizone.net) student portal. This
library is intended to be used as a self-hosted Go API or as an SDK in your Go application.

## Inspiration

Amizone is _the_ student portal for [Amity University](https://www.amity.edu/). It's indispensable for students to
access their grades, attendance, class schedule and other information. The catch: it's buggy, slow and goes down now and
then. Consequentially, students have made a slew of alternative apps and tools to access Amizone -- many of them mobile
apps, but also arguably better approaches like the excellent [monday-api][monday-api] bot by [@0xSaurabh][0xSaurabh],
which inspired me to work on this library.

**So why yet another tool?** Because I wanted a simple, easy-to-use and robust API client for Amizone and there was
none. There is no standard API-abstraction library for Amizone, so every tool has its own reverse-engineered way of
getting data from Amizone, some more complete than others, some more broken than others.

## Installation

The library can be installed either as an SDK to use in your own Go project or as a self-hosted API with the server
binary. With the latter, you would be able to use Swagger to generate SDKs for other languages in the near future.

### SDK
Install the library using `go get github.com/ditsuke/amizone-go`. Usage is well documented go docs, and docs
are due to be published soon.

### Server API
The server API is a RESTful API supported by a single go binary. It is intended to be used as a self-hosted API,
but I'll host a central deployment to make it easy to try out. To install locally, run:

```shell
go install github.com/ditsuke/amizone-go/amizone_api@latest
```

A docker image will be made available soon to make deployments easier.

## Contributing
I welcome contributions to the library. If you have a bug or feature request, please open an issue on the
[GitHub repo][github]. Contributing to project should be a great way to get started with Go development and learning
about the language, reverse-engineering, and other cool stuff.

[monday-api]: https://github.com/0xSaurabh/monday-api

[0xSaurabh]: https://github.com/0xSaurabh/

[github]: https://github.com/ditsuke/amizone-go
