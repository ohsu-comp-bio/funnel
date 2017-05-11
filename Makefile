ifndef GOPATH
$(error GOPATH is not set)
endif

export SHELL=/bin/bash
PATH := ${PATH}:${GOPATH}/bin
export PATH
PYTHONPATH := ${PYTHONPATH}:$(shell pwd)/python
export PYTHONPATH

PROTO_INC=-I ./ -I $(shell pwd)/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
GRPC_HTTP_MOD=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api

# Build the code
install: depends
	@go install github.com/ohsu-comp-bio/funnel

# Generate the protobuf/gRPC code
proto: depends
	@go get ./vendor/github.com/golang/protobuf/protoc-gen-go/
	@go get ./vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/
	@cd proto/tes && protoc \
		$(PROTO_INC) \
		--go_out=$(GRPC_HTTP_MOD),plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		tes.proto
	@cd proto/funnel && protoc \
		$(PROTO_INC) \
		-I ../tes \
		--go_out=$(GRPC_HTTP_MOD),Mtes.proto=github.com/ohsu-comp-bio/funnel/proto/tes,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		funnel.proto

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
	@pip install -q autopep8
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w -s
	@find . \( -path ./vendor -o -path ./webdash/node_modules -o -path ./venv -o -path ./.git \) -prune -o -type f -print | grep -E '.*\.py$$' | xargs autopep8 --in-place --aggressive --aggressive

# Run code style and other checks
lint:
	@pip install -q flake8
	@flake8 --exclude ./venv,./webdash,./vendor .
	@go get github.com/alecthomas/gometalinter
	@gometalinter --install > /dev/null
	@# TODO enable golint on funnel/cmd/termdash
	@gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --vendor \
	 -s proto --exclude 'cmd/examples/bundle.go' --exclude "cmd/termdash" --exclude 'webdash/web.go' ./...
	@gometalinter --disable-all --enable=vet --enable=gofmt --vendor ./cmd/termdash/...

# Run fast-running Go tests
go-test-short:
	@go test -short $(shell go list ./... | grep -v /vendor/)

# Run all Go tests
go-test:
	@go test $(shell go list ./... | grep -v /vendor/)

# Run all tests, including end-to-end Python tests
test:	go-test
	@docker build -t tes-wait -f tests/docker_files/tes-wait/Dockerfile tests/docker_files/tes-wait/
	@pip2.7 install -q -r tests/requirements.txt
	@nosetests-2.7 tests/

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
	@for GOOS in darwin linux; do \
		for GOARCH in 386 amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -o build/bin/funnel-$$GOOS-$$GOARCH .; \
		done; \
	done

# Upload a dev. release to GitHub
upload-dev-release:
	@go get github.com/aktau/github-release
	@if [ $$(git rev-parse --abbrev-ref HEAD) != 'master' ]; then \
		echo 'This command should only be run from the master branch'; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ]; then \
	  echo 'GITHUB_TOKEN env. var. is required but not set'; \
		exit 1; \
	fi
	@make gce-installer
	@mkdir -p build/dev-release
	@cp build/bin/* build/dev-release/
	@cp build/funnel-gce-image-installer build/dev-release
	@for GOOS in darwin linux; do \
		for GOARCH in 386 amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH \
				tar -C build/dev-release -czvf build/dev-release/funnel-$$GOOS-$$GOARCH.tar.gz funnel-$$GOOS-$$GOARCH; \
				rm build/dev-release/funnel-$$GOOS-$$GOARCH; \
		done; \
	done
	@for f in $$(ls -1 build/dev-release); do \
		github-release upload \
		-u ohsu-comp-bio \
		-r funnel \
		--name $$f \
		--tag dev \
		--replace \
		--file ./build/dev-release/$$f; \
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
	@mockery -dir scheduler/gce -name Client -print > scheduler/gce/mocks/Client_mock.go
	@mockery -dir scheduler/gce -name Wrapper -print > scheduler/gce/mocks/Wrapper_mock.go
	@mockery -dir server -name Database -print > server/mocks/Database_mock.go
	@mockery -dir scheduler -name Database -print > scheduler/mocks/Database_mock.go
	@mockery -dir scheduler -name Client -print > scheduler/mocks/Client_mock.go
	@mockery -dir proto/tes -name TaskServiceClient -print > proto/tes/mocks/TaskServiceClient_mock.go

# Bundle example task messages into Go code.
bundle-examples:
	@go-bindata -pkg examples -o cmd/examples/bundle.go examples

# Make everything usually needed to prepare for a pull request
full: proto install prune_deps add_deps tidy lint test web

# Build the website
website:
	@go get github.com/spf13/hugo
	hugo --source ./website
	# TODO there's more here
	# https://gohugo.io/tutorials/github-pages-blog/#deployment-via-gh-pages-branch

# Serve the Funnel website on localhost:1313
website-dev:
	@go get github.com/spf13/hugo
	hugo --source ./website -w server

# Remove build/development files.
clean:
	@rm -rf ./bin ./pkg ./test_tmp ./build ./buildtools

.PHONY: proto web website webdash
