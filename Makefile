# The commands used in this Makefile expect to be interpreted by bash.
SHELL := /bin/bash

TESTS=$(shell go list ./... | grep -v /vendor/ | grep -v github-release-notes)

PROTO_INC=-I ./ -I $(shell pwd)/util/proto/

git_commit := $(shell git rev-parse --short HEAD)
git_branch := $(shell git symbolic-ref -q --short HEAD)
git_upstream := $(shell git remote get-url $(shell git config branch.$(shell git symbolic-ref -q --short HEAD).remote) 2> /dev/null)
export GIT_BRANCH = $(git_branch)
export GIT_UPSTREAM = $(git_upstream)

export FUNNEL_VERSION=0.8.0

# LAST_PR_NUMBER is used by the release notes builder to generate notes
# based on pull requests (PR) up until the last release.
export LAST_PR_NUMBER = 546

VERSION_LDFLAGS=\
 -X "github.com/ohsu-comp-bio/funnel/version.BuildDate=$(shell date)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitCommit= $(git_commit)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitBranch=$(git_branch)" \
 -X "github.com/ohsu-comp-bio/funnel/version.GitUpstream=$(git_upstream)"

# Build the code
install: depends
	@touch version/version.go
	@go install -ldflags '$(VERSION_LDFLAGS)' github.com/ohsu-comp-bio/funnel

# Generate the protobuf/gRPC code
proto:
	@cd tes && protoc \
		$(PROTO_INC) \
		--go_out=plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		tes.proto
	@cd compute/scheduler && protoc \
		$(PROTO_INC) \
		--go_out=plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		scheduler.proto
	@cd events && protoc \
		$(PROTO_INC) \
		-I ../tes \
		--go_out=Mtes.proto=github.com/ohsu-comp-bio/funnel/tes,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		events.proto

# Update submodules and build code
depends:
	@git submodule update --init --recursive
	@go get github.com/golang/dep/cmd/dep
	@dep ensure

# Start API reference doc server
serve-doc:
	@go get golang.org/x/tools/cmd/godoc
	godoc --http=:6060

# Automatially update code formatting
tidy:
	@go get golang.org/x/tools/cmd/goimports
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v "\.pb\." | grep -v "web.go" | grep -E '.*\.go$$' | xargs goimports -w
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v "\.pb\." | grep -v "web.go" | grep -E '.*\.go$$' | xargs gofmt -w -s

# Run code style and other checks
lint:
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@# TODO enable golint on funnel/cmd/termdash
	@gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --enable=goimports --enable=misspell \
		--vendor \
		-e '.*bundle.go' -e ".*pb.go" -e ".*pb.gw.go" \
		-s "cmd/termdash" \
		-e 'webdash/web.go' -s 'funnel-work-dir' \
		./...
	@gometalinter --disable-all --enable=vet --enable=gofmt --enable=goimports --enable=misspell --vendor ./cmd/termdash/...

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
	@docker run -d --name funnel-mongodb-test -p 27000:27017 docker.io/mongo:3.5.13 > /dev/null

test-mongodb:
	@go test ./tests/core/ -funnel-config `pwd`/tests/mongo.config.yml
	@go test ./tests/scheduler/ -funnel-config `pwd`/tests/mongo.config.yml	

test-badger:
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
	@docker pull ohsucompbio/slurm
	@go test -timeout 120s ./tests/slurm -funnel-config `pwd`/tests/slurm.config.yml

test-gridengine:
	@docker pull ohsucompbio/gridengine
	@go test -timeout 120s ./tests/gridengine -funnel-config `pwd`/tests/gridengine.config.yml

test-pbs-torque:
	@docker pull ohsucompbio/pbs-torque
	@go test -timeout 120s ./tests/pbs -funnel-config `pwd`/tests/pbs.config.yml

test-amazon-s3:
	@go test ./tests/storage -funnel-config `pwd`/tests/s3.config.yml -run TestAmazonS3

start-generic-s3:
	@docker rm -f funnel-s3server > /dev/null 2>&1 || echo
	@docker run -d --name funnel-s3server -p 18888:8000 scality/s3server:mem-6018536a
	@docker rm -f funnel-minio > /dev/null 2>&1 || echo
	@docker run -d --name funnel-minio -p 9000:9000 -e "MINIO_ACCESS_KEY=fakekey" -e "MINIO_SECRET_KEY=fakesecret" -e "MINIO_REGION=us-east-1" minio/minio:RELEASE.2017-10-27T18-59-02Z server /data

test-generic-s3:
	@go test ./tests/storage -funnel-config `pwd`/tests/amazoncli-minio-s3.config.yml -run TestAmazonS3Storage
	@go test ./tests/storage -funnel-config `pwd`/tests/scality-s3.config.yml -run TestGenericS3Storage
	@go test ./tests/storage -funnel-config `pwd`/tests/minio-s3.config.yml -run TestGenericS3Storage
	@go test ./tests/storage -funnel-config `pwd`/tests/multi-s3.config.yml -run TestGenericS3Storage

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

webdash-install:
	@npm install --prefix ./webdash
	@go get -u github.com/jteeuwen/go-bindata/...

webdash-prep:
	@mkdir -p build/webdash
	@./webdash/node_modules/.bin/browserify webdash/app.js -o build/webdash/bundle.js
	@./webdash/node_modules/.bin/node-sass webdash/style.scss build/webdash/style.css
	@cp webdash/*.html build/webdash/
	@cp webdash/favicon.ico build/webdash/

# Build the web dashboard
webdash: webdash-prep
	@go-bindata -pkg webdash -prefix "build/" -o webdash/web.go build/webdash

# Build binaries for all OS/Architectures
snapshot: depends
	@goreleaser \
		--rm-dist \
		--snapshot

release:
	@go get github.com/buchanae/github-release-notes
	@goreleaser \
		--rm-dist \
		--release-notes <(github-release-notes -org ohsu-comp-bio -repo funnel -stop-at ${LAST_PR_NUMBER})

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

.PHONY: proto website docker webdash
