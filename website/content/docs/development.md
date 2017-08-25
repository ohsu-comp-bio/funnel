---
title: Development
menu:
  main:
    identifier: development
    weight: -30
    
---
# Development

Funnel uses:

- [Go][go] for the majority of the code.
- [Task Execution Schemas][tes] for task APIs.
- [Protobuf][protobuf] + [gRPC][grpc] for RPC communication.
- [gRPC Gateway][gateway] for HTTP communication.
- [Angular][angular] and [SASS][sass] for the web dashboard.
- [GNU Make][make] for development tasks.
- [Docker][docker] for executing task containers.
- [vendetta][vendetta] for Go dependency vendoring.
- and more.

## Prerequisites

These are the tools you'll need to install:

- [Go 1.8][go]
- [Make][make]
- [Docker][docker] (tested with v1.12, v1.13)
- [Protocol Buffers][protobuf] if making changes to the schema.
- [NodeJS][node] and [NPM][npm] for web dashboard development.

## Build

Most development tasks are run through `make`.

|Command|Description|
|---|---|
|`make`              | Build the code.
|`make test`         | Run both unit and end-to-end tests.
|`make test-short`   | Run only fast-running tests.
|`make test-verbose` | Run tests in verbose mode.
|`make test-backends`| Run end-to-end tests against dockerized HPC scheduler backends.
|`make proto`        | Regenerate code from protobuf schemas (requires protoc)
|`make tidy`         | Reformat code
|`make lint`         | Run code style and other checks.
|`make full`         | Run all steps needed to check the code before making a pull request.
|`make clean`        | Remove build/development files.
|`make add_deps`     | Add new vendored dependencies.
|`make prune_deps`   | Prune unused vendored dependencies.
|`make serve-doc`    | Serve API reference (godoc) docs on `localhost:6060`
|`make web`          | Build the web dashboard Javascript/CSS bundle.
|`make cross-compile`| Build binaries for all OS/Architectures.
|`make gce-installer`| Build the GCE image installer.
|`make gen-mocks`    | Generate mocks for testing.
|`make website-dev`   | Serve the Funnel website on localhost:1313
|`make upload-release`| Upload release binaries to GitHub.
|`make bundle-examples`| Bundle example task messages into Go code.

## Source

| Directory | Description |
|---|---|
|`cmd`              | Funnel command line interface.
|`config`           | Configuration parsing, loading, etc.
|`proto/tasklogger` | Internal, Funnel-specific protobuf/gRPC files for task state and log updates.
|`proto/tes`        | Generated GA4GH protobuf/gRPC files from [task-execution-schemas][tes].
|`proto/scheduler`  | Internal, Funnel-specific scheduler protobuf/gRPC files.
|`logger`           | Logging.
|`scheduler`        | Basic scheduling/scaling logic and backends.
|`server`           | Database and server implementing the [TES API][tes] and Scheduler RPC.
|`storage`          | Filesystem support, e.g. local, Google Cloud Storage, S3, etc.
|`worker`           | Worker process: task runner, docker, file mapper, etc.
|`webdash`          | Javascript, CSS, HTML for web dashboard.

## Go Tests

Run all tests: `make go-test`   
Run the scheduler tests: `go test ./scheduler/...`  
Run the worker tests with "Cancel" in the name: `go test ./worker -run Cancel`  

You get the idea. See the `go test` docs for more.

## Mocking

The [testify][testify] and [mockery][mockery] tools are used to generate and use
mock interfaces in test code, for example, to mock the Google Cloud APIs.

## Vendoring

Go dependencies are vendored under /vendor.  
Don't manually add new submodules, use `make add_deps`.

## Submodules

Funnel has git submodules. The Makefile usually handles this for you, but if needed,
`git submodule update --init --recursive` will get all the submodules.

## Release Process

This list is a work in progress:

- edit Makefile to update version
- set up GitHub API auth using a token (see Makefile)
- run `make upload-release`
- edit website/content/install.md to replace download links
- edit website/layouts/index.html to replace the download button text
- release the website

Does that seem too manual and error-prone to you? You're right! See: https://github.com/ohsu-comp-bio/funnel/issues/186

[go]: https://golang.org
[angular]: https://angularjs.org/
[protobuf]: https://github.com/google/protobuf
[grpc]: http://www.grpc.io/
[sass]: http://sass-lang.com/
[make]: https://www.gnu.org/software/make/
[docker]: https://docker.io
[python]: https://www.python.org/
[vendetta]: https://github.com/dpw/vendetta
[node]: https://nodejs.org
[npm]: https://www.npmjs.com/
[gateway]: https://github.com/grpc-ecosystem/grpc-gateway
[tes]: https://github.com/ga4gh/task-execution-schemas
[testify]: https://github.com/stretchr/testify
[mockery]: https://github.com/vektra/mockery
