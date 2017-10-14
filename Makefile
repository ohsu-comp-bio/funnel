ifndef GOPATH
$(error GOPATH is not set)
endif

VERSION = 0.2.0
TESTS=$(shell go list ./... | grep -v /vendor/)

export SHELL=/bin/bash
PATH := ${PATH}:${GOPATH}/bin
export PATH

PROTO_INC=-I ./  -I $(shell pwd)/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis

V=github.com/ohsu-comp-bio/funnel/version
VERSION_LDFLAGS=\
 -X "$(V).BuildDate=$(shell date)" \
 -X "$(V).GitCommit=$(shell git rev-parse --short HEAD)" \
 -X "$(V).GitBranch=$(shell git symbolic-ref -q --short HEAD)" \
 -X "$(V).GitUpstream=$(shell git remote get-url $(shell git config branch.$(shell git symbolic-ref -q --short HEAD).remote) 2> /dev/null)" \
 -X "$(V).Version=$(shell git describe --tags --long --dirty)"

# Build the code
install: depends
	@touch version/version.go
	@go install -ldflags '$(VERSION_LDFLAGS)' github.com/ohsu-comp-bio/funnel

# Generate the protobuf/gRPC code
proto:
	@cd proto/tes && protoc \
		$(PROTO_INC) \
		--go_out=plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		tes.proto
	@cd proto/scheduler && protoc \
		$(PROTO_INC) \
		--go_out=plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		scheduler.proto
	@cd events && protoc \
		$(PROTO_INC) \
		-I ../proto/tes \
		-I $(shell pwd)/vendor/github.com/golang/protobuf/ptypes/struct/ \
		-I $(shell pwd)/vendor/github.com/golang/protobuf/ptypes/timestamp/ \
		--go_out=Mtes.proto=github.com/ohsu-comp-bio/funnel/proto/tes,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		events.proto

# Update submodules and build code
depends:
	@git submodule update --init --recursive
	@go get -d github.com/ohsu-comp-bio/funnel

# Start API reference doc server
serve-doc:
	godoc --http=:6060

# Add new vendored dependencies
add_deps:
	@go get github.com/dpw/vendetta
	@vendetta ./

# Prune unused vendored dependencies
prune_deps:
	@go get github.com/dpw/vendetta
	@vendetta -p ./

# Automatially update code formatting
tidy:
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v "\.pb\." | grep -v "web.go" | grep -E '.*\.go$$' | xargs gofmt -w -s

# Run code style and other checks
lint:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@# TODO enable golint on funnel/cmd/termdash
	@gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --enable=misspell \
		--vendor \
		-e '.*bundle.go' -e ".*pb.go" -e ".*pb.gw.go" \
		-s "cmd/termdash" \
		-e 'webdash/web.go' -s 'funnel-work-dir' \
		./...
	@gometalinter --disable-all --enable=vet --enable=gofmt --enable=misspell --vendor ./cmd/termdash/...

# Run fast-running Go tests
test-short:
	@go test -short $(TESTS)

# Run all tests
test:
	@go test $(TESTS)

test-verbose:
	@go test -v $(TESTS)

# Run backend tests
test-backends:
	@go test -timeout 120s ./tests/e2e/slurm -run-test
	@go test -timeout 120s ./tests/e2e/gridengine -run-test
	@go test -timeout 120s ./tests/e2e/htcondor -run-test
	@go test -timeout 120s ./tests/e2e/pbs -run-test

# Run s3 tests
test-s3:
	@go test ./tests/e2e/s3 -run-test

# Tests meant to run in an OpenStack environment
test-openstack:
	@go test ./tests/e2e/openstack -openstack-e2e-config ${FUNNEL_OPENSTACK_TEST_CONFIG}

# Build the web dashboard
webdash:
	@mkdir -p build/webdash
	@npm install --prefix ./webdash
	@./webdash/node_modules/.bin/browserify webdash/app.js -o build/webdash/bundle.js
	@./webdash/node_modules/.bin/node-sass webdash/style.scss build/webdash/style.css
	@cp webdash/*.html build/webdash/
	@go get -u github.com/jteeuwen/go-bindata/...
	@go-bindata -pkg webdash -prefix "build/" -o webdash/web.go build/webdash

# Build binaries for all OS/Architectures
cross-compile: depends
	@echo '=== Cross compiling... ==='
	@for GOOS in darwin linux; do \
		for GOARCH in amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -a \
				-ldflags '$(VERSION_LDFLAGS)' \
				-o build/bin/funnel-$$GOOS-$$GOARCH .; \
		done; \
	done

# Upload a release to GitHub
upload-release:
	#
	# NOTE! Making a release requires manual steps.
	# See: website/content/docs/development.md
	@go get github.com/aktau/github-release
	@if [ $$(git rev-parse --abbrev-ref HEAD) != 'master' ]; then \
		echo 'This command should only be run from the master branch'; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo 'GITHUB_TOKEN is required but not set. Generate one in your GitHub settings at https://github.com/settings/tokens and set it to an environment variable with `export GITHUB_TOKEN=123456...`'; \
		exit 1; \
	fi
	@make gce-installer
	@mkdir -p build/release
	@cp build/bin/* build/release/
	@cp build/funnel-gce-image-installer build/release
	-github-release release \
		-u ohsu-comp-bio \
		-r funnel \
		--tag $(VERSION) \
		--name $(VERSION)
	for f in $$(ls -1 build/release); do \
		github-release upload \
		-u ohsu-comp-bio \
		-r funnel \
		--name $$f \
		--tag $(VERSION) \
		--replace \
		--file ./build/release/$$f; \
	done

# Build the GCE image installer
gce-installer: cross-compile
	@mkdir -p build/gce-installer
	@cp deployments/gce/bundle/* build/gce-installer/
	@cp build/bin/funnel-linux-amd64 build/gce-installer/funnel
	@cd build && \
		../deployments/gce/make-installer.sh -c gce-installer && \
		mv bundle.run funnel-gce-image-installer && \
		cd ..

# Generate mocks for testing.
gen-mocks:
	@go get github.com/vektra/mockery/...
	@mockery -dir compute/scheduler -name Database -print > compute/scheduler/mocks/Database_mock.go
	@mockery -dir compute/scheduler -name Client -print > compute/scheduler/mocks/Client_mock.go
	@mockery -dir compute/gce -name Client -print > compute/gce/mocks/Client_mock.go
	@mockery -dir compute/gce -name Wrapper -print > compute/gce/mocks/Wrapper_mock.go
	@mockery -dir server -name Database -print > server/mocks/Database_mock.go

# Bundle example task messages into Go code.
bundle-examples:
	@go-bindata -pkg examples -o examples/bundle.go $(shell find examples/ -name '*.json')
	@go-bindata -pkg config -o config/bundle.go $(shell find config/ -name '*.txt' -o -name '*.yaml')

# Make everything usually needed to prepare for a pull request
full: proto install prune_deps add_deps tidy lint test website webdash

# Build the website
website:
	@find ./config -name '*.txt' -o -name '*.yaml' | xargs -I % cp % ./website/static/funnel-config-examples/
	@go get github.com/spf13/hugo
	hugo --source ./website
	#
	# NOTE! release the website requires manual steps.
	# TODO there's more here
	# https://gohugo.io/tutorials/github-pages-blog/#deployment-via-gh-pages-branch

# Serve the Funnel website on localhost:1313
website-dev:
	@find ./config -name '*.txt' -o -name '*.yaml' -exec cp {} website/static/funnel-config-examples/ \;
	@go get github.com/spf13/hugo
	hugo --source ./website -w server

# Build docker image.
docker: cross-compile
	mkdir -p build/docker
	cp build/bin/funnel-linux-amd64 build/docker/funnel
	cp docker/* build/docker/
	cd build/docker/ && docker build -t funnel .

# Remove build/development files.
clean:
	@rm -rf ./bin ./pkg ./test_tmp ./build ./buildtools

.PHONY: proto website docker webdash
