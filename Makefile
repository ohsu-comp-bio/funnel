# The commands used in this Makefile expect to be interpreted by bash.
SHELL := /bin/bash

TESTS=$(shell go list ./... | grep -v /vendor/ | grep -v github-release-notes)

PROTO_INC=-I ./ -I $(shell pwd)/util/proto/

git_commit := $(shell git rev-parse --short HEAD)
git_branch := $(shell git symbolic-ref -q --short HEAD)
git_upstream := $(shell git remote get-url $(shell git config branch.$(shell git symbolic-ref -q --short HEAD).remote) 2> /dev/null)
export GIT_BRANCH = $(git_branch)
export GIT_UPSTREAM = $(git_upstream)

export FUNNEL_VERSION=0.11.0

# LAST_PR_NUMBER is used by the release notes builder to generate notes
# based on pull requests (PR) up until the last release.
export LAST_PR_NUMBER = 716

VERSION_LDFLAGS=\
 -X "github.com/ohsu-comp-bio/funnel/version.BuildDate=$(shell date)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitCommit= $(git_commit)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitBranch=$(git_branch)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitUpstream=$(git_upstream)"

export CGO_ENABLED=0

# Build the code
install:
	@touch version/version.go
	@go install -ldflags '$(VERSION_LDFLAGS)' .

# Build the code
build:
	@touch version/version.go
	@go build -ldflags '$(VERSION_LDFLAGS)' -buildvcs=false .

# Build an unoptimized version of the code for use during debugging 
# https://go.dev/doc/gdb
debug:
	@go install -gcflags=all="-N -l"
	@funnel server run

# Generate the protobuf/gRPC code
proto:
	@cd tes && protoc \
		$(PROTO_INC) \
		--go_out ./ \
  		--go_opt paths=source_relative \
		--go-grpc_out ./ \
		--go-grpc_opt paths=source_relative \
		--grpc-gateway_out ./ \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		tes.proto
	@cd compute/scheduler && protoc \
		$(PROTO_INC) \
		--go_out ./ \
  		--go_opt paths=source_relative \
		--go-grpc_out ./ \
		--go-grpc_opt paths=source_relative \
		--grpc-gateway_out ./ \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		scheduler.proto
	@cd events && protoc \
		$(PROTO_INC) \
		-I ../tes \
		--go_out ./ \
  		--go_opt paths=source_relative \
		--go-grpc_out ./ \
		--go-grpc_opt paths=source_relative \
		--grpc-gateway_out ./ \
		--grpc-gateway_opt logtostderr=true \
		--grpc-gateway_opt paths=source_relative \
		events.proto


proto-depends:
	@git submodule update --init --recursive
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.11.1
	@go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.11.1
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@go install github.com/ckaznocha/protoc-gen-lint@v0.2.4

# Start API reference doc server
serve-doc:
	@go get golang.org/x/tools/cmd/godoc
	godoc --http=:6060

# Automatially update code formatting
tidy:
	@go get golang.org/x/tools/cmd/goimports
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v "\.pb\." | grep -v "web.go" | grep -E '.*\.go$$' | xargs goimports -w
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v "\.pb\." | grep -v "web.go" | grep -E '.*\.go$$' | xargs gofmt -w -s

lint-depends:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.1
	go install golang.org/x/tools/cmd/goimports

# Run code style and other checks
lint:
	@golangci-lint run --timeout 3m --disable-all --enable=vet --enable=golint --enable=gofmt --enable=goimports --enable=misspell \
		--skip-dirs "vendor" \
		--skip-dirs "webdash" \
		--skip-dirs "cmd/webdash" \
		--skip-dirs "funnel-work-dir" \
		-e '.*bundle.go' -e ".*pb.go" -e ".*pb.gw.go" \
		./...
	@golangci-lint run --timeout 3m --disable-all --enable=vet --enable=gofmt --enable=goimports --enable=misspell ./cmd/termdash/...

# Run all tests
test:
	@go test $(TESTS)

test-verbose:
	@go test -v $(TESTS)

start-elasticsearch:
	@docker rm -f funnel-es-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-es-test -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "xpack.security.enabled=false" docker.elastic.co/elasticsearch/elasticsearch:5.6.3 > /dev/null

test-elasticsearch:
	@go test ./tests/core/ -funnel-config `pwd`/tests/elastic.config.yml
	@go test ./tests/scheduler/ -funnel-config `pwd`/tests/elastic.config.yml

start-mongodb:
	@docker rm -f funnel-mongodb-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-mongodb-test -p 27000:27017 docker.io/mongo > /dev/null

test-mongodb:
	@go version
	@go test ./tests/core/ --funnel-config `pwd`/tests/mongo.config.yml
	@go test ./tests/scheduler/ --funnel-config `pwd`/tests/mongo.config.yml

test-badger:
	@go version
	@go test ./tests/core/ -funnel-config `pwd`/tests/badger.config.yml

start-dynamodb:
	@docker rm -f funnel-dynamodb-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-dynamodb-test -p 18000:8000 docker.io/dwmkerr/dynamodb:38 -sharedDb > /dev/null

test-dynamodb:
	@go test ./tests/core/ -funnel-config `pwd`/tests/dynamo.config.yml

start-datastore:
	@docker rm -f funnel-datastore-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-datastore-test -p 12432:12432 google/cloud-sdk:latest gcloud beta emulators datastore start --no-store-on-disk --host-port 0.0.0.0:12432 --project funnel-test

test-datastore:
	DATASTORE_EMULATOR_HOST=localhost:12432 \
	go test ./tests/core/ -funnel-config `pwd`/tests/datastore.config.yml

start-kafka:
	@docker rm -f funnel-kafka > /dev/null 2>&1 || echo
	@docker run -d --name funnel-kafka -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST="localhost" --env ADVERTISED_PORT=9092 spotify/kafka

test-kafka:
	@go test ./tests/kafka/ -funnel-config `pwd`/tests/kafka.config.yml

test-htcondor:
	@docker pull ohsucompbio/htcondor
	@go test -timeout 120s ./tests/htcondor -funnel-config `pwd`/tests/htcondor.config.yml

test-slurm:
	@go version
	@docker pull quay.io/ohsu-comp-bio/slurm
	@go test -timeout 120s ./tests/slurm -funnel-config `pwd`/tests/slurm.config.yml

test-gridengine:
	@docker pull ohsucompbio/gridengine
	@go test -timeout 120s ./tests/gridengine -funnel-config `pwd`/tests/gridengine.config.yml

test-pbs-torque:
	@docker pull ohsucompbio/pbs-torque
	@go test -timeout 120s ./tests/pbs -funnel-config `pwd`/tests/pbs.config.yml

test-amazon-s3:
	@go test -v ./tests/storage -funnel-config `pwd`/tests/s3.config.yml -run TestAmazonS3

start-generic-s3:
	@docker rm -f funnel-s3server > /dev/null 2>&1 || echo
	@docker run -d --name funnel-s3server -p 18888:8000 -e REMOTE_MANAGEMENT_DISABLE=1 zenko/cloudserver
	@docker rm -f funnel-minio > /dev/null 2>&1 || echo
	@docker run -d --name funnel-minio -p 9000:9000 -e "MINIO_ACCESS_KEY=fakekey" -e "MINIO_SECRET_KEY=fakesecret" -e "MINIO_REGION=us-east-1" minio/minio server /data

test-generic-s3:
	@go test -v ./tests/storage -funnel-config `pwd`/tests/amazoncli-minio-s3.config.yml -run TestAmazonS3Storage
	@go test -v ./tests/storage -funnel-config `pwd`/tests/scality-s3.config.yml -run TestGenericS3Storage
	@go test -v ./tests/storage -funnel-config `pwd`/tests/minio-s3.config.yml -run TestGenericS3Storage
	@go test -v ./tests/storage -funnel-config `pwd`/tests/multi-s3.config.yml -run TestGenericS3Storage

test-gs:
	@go test ./tests/storage -run TestGoogleStorage -funnel-config `pwd`/tests/gs.config.yml ${GCE_PROJECT_ID}

test-swift:
	@go test ./tests/storage -funnel-config `pwd`/tests/swift.config.yml -run TestSwiftStorage

start-pubsub:
	@docker rm -f funnel-pubsub-test > /dev/null 2>&1 || echo
	@docker run -d --name funnel-pubsub-test -p 8085:8085 google/cloud-sdk:latest gcloud beta emulators pubsub start --project funnel-test --host-port 0.0.0.0:8085

test-pubsub:
	@PUBSUB_EMULATOR_HOST=localhost:8085 \
	go test ./tests/pubsub/ -funnel-config `pwd`/tests/pubsub.config.yml

start-ftp:
	@cd tests/ftp-test-server/ && ./start-server.sh

test-ftp:
	@go test -v ./tests/storage -run TestFTPStorage -funnel-config `pwd`/tests/ftp.config.yml

# Build the web dashboard
webdash:
	@go get -u github.com/go-bindata/go-bindata/...
	@cd webdash && yarn build
	@go-bindata -pkg webdash -prefix "webdash/build" -o webdash/web.go webdash/build/...

# Build binaries for all OS/Architectures
snapshot: release-dep
	@goreleaser \
		--clean \
		--snapshot

# build a docker container locally
docker:
	docker build -t ohsucompbio/funnel:latest ./

# build a docker container that supports docker-in-docker
docker-dind:
	docker build -t ohsucompbio/funnel-dind:latest -f Dockerfile.dind .

# build a docker container that supports rootless docker-in-docker
docker-dind-rootless:
	docker build -t ohsucompbio/funnel-dind-rootless:latest -f Dockerfile.dind-rootless .

# Create a release on Github using GoReleaser 
release:
	@go get github.com/buchanae/github-release-notes
	@goreleaser \
		--clean \
		--release-notes <(github-release-notes -org ohsu-comp-bio -repo funnel -stop-at ${LAST_PR_NUMBER})

# Install dependencies for release
release-dep:
	@go get github.com/goreleaser/goreleaser
	@go get github.com/buchanae/github-release-notes

# Generate mocks for testing.
gen-mocks:
	@go get github.com/vektra/mockery/...
	@mockery -dir compute/scheduler -name Client -inpkg -output compute/scheduler
	@mockery -dir compute/scheduler -name SchedulerServiceServer -inpkg -output compute/scheduler

# Bundle example task messages into Go code.
bundle-examples:
	@go-bindata -pkg examples -o examples/internal/bundle.go $(shell find examples/ -name '*.json')
	@go-bindata -pkg config -o config/internal/bundle.go $(shell find config/ -name '*.txt' -o -name '*.yaml')
	@gofmt -w -s examples/internal/bundle.go config/internal/bundle.go

# Make everything usually needed to prepare for a pull request
full: proto install tidy lint test website webdash

# Build the website
website:
	@cp ./config/*.txt ./website/static/funnel-config-examples/
	@cp ./config/default-config.yaml ./website/static/funnel-config-examples/
	hugo --source ./website

# Serve the Funnel website on localhost:1313
website-dev:
	@cp ./config/*.txt ./website/static/funnel-config-examples/
	@cp ./config/default-config.yaml ./website/static/funnel-config-examples/
	hugo --source ./website -w server

# Remove build/development files.
clean:
	@rm -rf ./bin ./pkg ./test_tmp ./build ./buildtools

.PHONY: proto website docker webdash build debug
