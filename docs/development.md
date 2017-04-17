# Development

Funnel...
- is written in Go
- uses Protobuf + gRPC for communication
- uses Angular and SASS for the web dashboard
- uses make for development tasks
- executes tasks via Docker
- has tests, examples, and a client in Python
- vendors dependencies with [vendetta](github.com/dpw/vendetta)
- and more.

## Prerequisites

- [Go 1.7](https://golang.org/)
- Make
- [Docker](https://docker.io/) (tested with v1.12)
- [Protocol Buffers](https://github.com/google/protobuf) if making changes to the schema.
- [NodeJS](https://nodejs.org) and [npm](https://www.npmjs.com/) for web dashboard development.

## GOPATH

Funnel isn't go-gettable and the layout doesn't work with Go cleanly (yet).

Run ``export GOPATH=$GOPATH:`pwd` `` from the repo root to get the funnel code in your GOPATH.

## Submodules

Funnel has git submodules. Make sure to run `git submodule update --init --recursive`. Running `make` will do this for you.

## Build

Most commands run through `make`. Binaries are built to `./bin`.
Unfortunately this project isn't "go get-able" yet.

- `make` builds the code
- `make test` runs the test suite
- `make proto` regenerates code from protobuf schemas (requires protoc)
- `make tidy` reformats code
- `make lint` checks code style and other 
- `make add_deps` uses [vendetta](github.com/dpw/vendetta) to vendor Go dependencies
- `make serve-doc` runs a godoc server on `localhost:6060`
- `make web` builds the web dashboard CSS.

There are probably other commands. Check out the [Makefile](../Makefile).

## Source

- `src/funnel`: majority of the code.
  - `cmd`: funnel command line interface
  - `config`: configuration parsing/loading
  - `proto/tes`: generated GA4GH protobuf/gRPC files from [task-execution-schemas](../task-execution-schemas/proto/)
  - `proto/funnel`: the internal, Funnel-specific protobuf/gRPC files
  - `logger`: custom logging code
  - `scheduler`: scheduler logic and scheduler backends
  - `server`: database and server implementing the TES API
  - `storage`: filesystem support, used by worker during upload/download, e.g. local, Google Cloud Storage, S3, etc.
  - `worker`: worker process and state management, task runner, docker command executor, file mapper, etc.
- `bin/funnel`: funnel CLI binary

- `web`: javascript/css/html for web dashboard

## Go Tests

Useful testing commands (first, see the GOPATH section above):

Run all tests: `go test funnel/...`

Run the scheduler tests: `go test funnel/scheduler/...`

Run the worker tests matching "\*Cancel\*": `go test funnel/worker -run Cancel`

You get the idea. See to `go test` docs for more.

## Mocking

There are some helpful mocks in the code. The [testify](https://github.com/stretchr/testify) and [mockery](https://github.com/vektra/mockery) tools have been useful.

There's also a mock server for testing in [src/tes/server/mocks/server.go](./src/tes/server/mocks/server.go).

## Python Tests

There are integration tests written in python. These are heavyweight integration tests which start funnel server and workers processes and run docker containers. Usually these are run with `make test`.

## Vendoring

Go dependencies are vendored using [vendetta](github.com/dpw/vendetta) under src/vendor. Don't manually add new submodules, use vendetta! It's easy to get the vendoring wrong and vendetta makes it easy. Run `make add_deps`.
