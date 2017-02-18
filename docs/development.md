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

## Submodules

Funnel has git submodules. Make sure to run `git submodule update --init --recursive`. Running `make` will do this for you.

## Build

Most commands run through `make`. Binaries are built to `./bin`.
Unfortunately this project isn't "go get-able" yet.

`make` builds the code
`make test` runs the test suite
`make proto_build` regenerates code from protobuf schemas (requires protoc)
`make tidy` reformats code
`make lint` checks code style and other 
`make add_deps` uses [vendetta](github.com/dpw/vendetta) to vendor Go dependencies
`make serve-doc` runs a godoc server on `localhost:6060`

There are probably other commands. Check out the Makefile.

## Source

- `src/tes`: majority of the code. The subdir names should be fairly obvious.
- `proto`: the internal, Funnel-specific protobuf + gRPC schemas.
- `task-execution-schemas/proto`: the GA4GH Protobuf + gRPC schemas.
- `share`: javascript/css/html for web dashboard.


## Vendoring

Go dependencies are vendored using [vendetta](github.com/dpw/vendetta) under src/vendor. Don't manually add new submodules, use vendetta! It's easy to get the vendoring wrong and vendetta makes it easy. Run `make add_deps`.
