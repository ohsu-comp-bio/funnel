---
title: Funnel Developers

menu:
  main:
    parent: Development
    weight: 30
---

# Developers

This page contains a rough collection of notes for people wanting to build Funnel from source and/or edit the code.

### Building the Funnel source

1. Install [Go 1.8+][go]. Check the version with `go version`.
1. Ensure GOPATH is set. See [the docs][gopath] for help. Also, you probably want to add `$GOPATH/bin` to your `PATH`.
1. Run `go get github.com/ohsu-comp-bio/funnel`
1. Funnel is now downloaded and installed. Try `funnel version`.
1. `cd $GOPATH/src/github.com/ohsu-comp-bio/funnel`
1. Now you're in the Funnel repo. You can edit the code an rerun the `go get` command above, or check out the Makefile for a list of useful build commands, such as `make install`.

### Developer Tools

A Funnel development environment includes:

- [Go 1.8+][go] for the majority of the code.
- [Task Execution Schemas][tes] for task APIs.
- [Protobuf][protobuf] + [gRPC][grpc] for RPC communication.
- [gRPC Gateway][gateway] for HTTP communication.
- [Angular][angular] and [SASS][sass] for the web dashboard.
- [GNU Make][make] for development tasks.
- [Docker][docker] for executing task containers (tested with v1.12, v1.13).
- [vendetta][vendetta] for Go dependency vendoring.
- [Make][make] for development/build commands.
- [NodeJS][node] and [NPM][npm] for web dashboard development.

### Makefile

Most development tasks are run through `make` commands, including build, release, testing, website docs, lint, tidy, webdash dev, and more.  See the [Makefile](https://github.com/ohsu-comp-bio/funnel/blob/master/Makefile) for an up-to-date list of commands.

### Go Tests

Run all tests: `make test`   
Run the worker tests: `go test ./worker/...`  
Run the worker tests with "Cancel" in the name: `go test ./worker -run Cancel`  

You get the idea. See the `go test` docs for more.

### Mocking

The [testify][testify] and [mockery][mockery] tools are used to generate and use
mock interfaces in test code, for example, to mock the Google Cloud APIs.

### Vendoring

Go dependencies are vendored under /vendor. `make add_deps` and `make prune_deps` help manage new dependencies.

### Submodules

Funnel has git submodules. The Makefile usually handles this for you, but if needed,
`git submodule update --init --recursive` will get all the submodules.

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
[gopath]: https://golang.org/doc/code.html#GOPATH
