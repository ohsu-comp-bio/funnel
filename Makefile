ifndef GOPATH
$(error GOPATH is not set)
endif

VERSION = 0.4.0
TESTS=$(shell go list ./... | grep -v /vendor/)
CONFIGDIR=$(shell pwd)/tests

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
 -X "$(V).Version=$(VERSION)"

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

test-travis:
ifeq ($(BUILD_TYPE), lint)
	make lint
else
	make test-${BUILD_TYPE}
endif

# Run all tests
test:
	@go test $(TESTS)

test-verbose:
	@go test -v $(TESTS)

start-elasticsearch:
	@docker rm -f funnel-es-test  > /dev/null 2>&1 || echo
	@docker run -d --name funnel-es-test -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=false" docker.elastic.co/elasticsearch/elasticsearch:5.6.3 > /dev/null

test-elasticsearch:
	@go test ./tests/core/ -funnel-config $(CONFIGDIR)/elastic.config.yml
	@go test ./tests/scheduler/ -funnel-config $(CONFIGDIR)/elastic.config.yml

start-mongodb:
	@docker rm -f funnel-mongodb-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-mongodb-test -p 27000:27017 docker.io/mongo:3.5.13 > /dev/null

test-mongodb:
	@go test ./tests/core/ -funnel-config $(CONFIGDIR)/mongo.config.yml
	@go test ./tests/scheduler/ -funnel-config $(CONFIGDIR)/mongo.config.yml	

start-dynamodb:
	@docker rm -f funnel-dynamodb-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-dynamodb-test -p 8000:8000 docker.io/dwmkerr/dynamodb:38 > /dev/null

test-dynamodb:
	@go test ./tests/core/ -funnel-config $(CONFIGDIR)/dynamo.config.yml

start-kafka:
	@docker rm -f funnel-kafka > /dev/null 2>&1 || echo
	@docker run -d --name funnel-kafka -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST="localhost" --env ADVERTISED_PORT=9092 spotify/kafka

test-kafka:
	@go test ./tests/kafka/

test-htcondor:
	@docker pull ohsucompbio/htcondor
	@go test -timeout 120s ./tests/htcondor -funnel-config $(CONFIGDIR)/htcondor.config.yml

test-slurm:
	@docker pull ohsucompbio/slurm
	@go test -timeout 120s ./tests/slurm -funnel-config $(CONFIGDIR)/slurm.config.yml

test-gridengine:
	@docker pull ohsucompbio/gridengine
	@go test -timeout 120s ./tests/gridengine -funnel-config $(CONFIGDIR)/gridengine.config.yml

test-pbs-torque:
	@docker pull ohsucompbio/pbs-torque
	@go test -timeout 120s ./tests/pbs -funnel-config $(CONFIGDIR)/pbs.config.yml

# Run s3 tests
test-s3:
	@go test ./tests/storage -funnel-config $(CONFIGDIR)/s3.config.yml

# Tests meant to run in an OpenStack environment
test-swift:
	@go test ./tests/storage -funnel-config $(CONFIGDIR)/swift.config.yml

webdash-install:
	@npm install --prefix ./webdash
	@go get -u github.com/jteeuwen/go-bindata/...

webdash-prep:
	@mkdir -p build/webdash
	@./webdash/node_modules/.bin/browserify webdash/app.js -o build/webdash/bundle.js
	@./webdash/node_modules/.bin/node-sass webdash/style.scss build/webdash/style.css
	@cp webdash/*.html build/webdash/
	@cp webdash/favicon.ico build/webdash/

webdash-debug: webdash-prep
	@go-bindata -debug -pkg webdash -prefix "build/" -o webdash/web.go build/webdash

# Build the web dashboard
webdash: webdash-prep
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

clean-release:
	rm -rf ./build/release

# Upload a release to GitHub
upload-release: clean-release cross-compile
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
	-github-release release \
		-u ohsu-comp-bio \
		-r funnel \
		--tag $(VERSION) \
		--name $(VERSION)
	for f in $$(ls -1 build/bin); do \
		mkdir -p build/release/$$f-$(VERSION); \
		cp build/bin/$$f build/release/$$f-$(VERSION)/funnel; \
		tar -C build/release/$$f-$(VERSION) -czf build/release/$$f-$(VERSION).tar.gz .; \
		github-release upload \
		-u ohsu-comp-bio \
		-r funnel \
		--name $$f-$(VERSION).tar.gz \
		--tag $(VERSION) \
		--replace \
		--file ./build/release/$$f-$(VERSION).tar.gz; \
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
	@go get github.com/spf13/hugo
	hugo --source ./website

# Serve the Funnel website on localhost:1313
website-dev:
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
