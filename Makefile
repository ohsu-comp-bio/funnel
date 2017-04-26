GOPATH := $(shell pwd)/build:$(shell pwd)
export SHELL=/bin/bash
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin:$(shell pwd)/build/bin
export PATH
PYTHONPATH := ${PYTHONPATH}:$(shell pwd)/python
export PYTHONPATH

PROTO_INC=-I ./ -I $(GOPATH)/src/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
GRPC_HTTP_MOD=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api

install: depends
	go install funnel

proto: depends
	@go get ./src/vendor/github.com/golang/protobuf/protoc-gen-go/
	@go get ./src/vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/
	@cd src/funnel/proto/tes && protoc \
		$(PROTO_INC) \
		--go_out=$(GRPC_HTTP_MOD),plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		tes.proto
	@cd src/funnel/proto/funnel && protoc \
		$(PROTO_INC) \
		-I ../tes \
		--go_out=$(GRPC_HTTP_MOD),Mtes.proto=funnel/proto/tes,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		funnel.proto

depends:
	git submodule update --init --recursive
	go get -d funnel

serve-doc:
	godoc --http=:6060

add_deps: 
	go get github.com/dpw/vendetta
	./build/bin/vendetta src/

prune_deps:
	go get github.com/dpw/vendetta
	./build/bin/vendetta -p src/

tidy:
	pip install -q autopep8
	@find ./src/funnel* -type f | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w -s
	@find ./* -type f | grep -E '.*\.py$$' | grep -v "/venv/" | grep -v "/web/node" | xargs autopep8 --in-place --aggressive --aggressive

lint:
	pip install -q flake8
	flake8 --exclude ./venv,./web .
	go get github.com/alecthomas/gometalinter
	./build/bin/gometalinter --install > /dev/null
	./build/bin/gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --vendor -s ga4gh -s proto -s web --exclude 'examples/bundle.go' ./src/funnel/...

go-test-short:
	go test -short funnel/...

go-test:
	go test funnel/...

test:	go-test
	docker build -t tes-wait -f tests/docker_files/tes-wait/Dockerfile tests/docker_files/tes-wait/
	pip2.7 install -q -r tests/requirements.txt
	nosetests-2.7 tests/

web:
	mkdir -p build/web
	npm install --prefix ./web
	./web/node_modules/.bin/browserify web/app.js -o build/web/bundle.js
	./web/node_modules/.bin/node-sass web/style.scss build/web/style.css
	cp web/*.html build/web/
	go get -u github.com/jteeuwen/go-bindata/...
	go-bindata -pkg web -prefix "build/" -o src/funnel/web/web.go build/web

cross-compile: depends
	for GOOS in darwin linux; do \
		for GOARCH in 386 amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH go build -o ./bin/funnel-$$GOOS-$$GOARCH funnel; \
		done; \
	done

upload-dev-release:
	go get github.com/aktau/github-release
	@if [ $$(git rev-parse --abbrev-ref HEAD) != 'master' ]; then \
		echo 'This command should only be run from the master branch'; \
		exit 1; \
	fi
	@if [ -z "$$GITHUB_TOKEN" ]; then \
	  echo 'GITHUB_TOKEN env. var. is required but not set'; \
		exit 1; \
	fi
	make gce-installer
	mkdir -p build/dev-release
	cp bin/* build/dev-release/
	cp build/funnel-gce-image-installer build/dev-release
	for GOOS in darwin linux; do \
		for GOARCH in 386 amd64; do \
			GOOS=$$GOOS GOARCH=$$GOARCH \
				tar -C build/dev-release -czvf build/dev-release/funnel-$$GOOS-$$GOARCH.tar.gz funnel-$$GOOS-$$GOARCH; \
				rm build/dev-release/funnel-$$GOOS-$$GOARCH; \
		done; \
	done
	for f in $$(ls -1 build/dev-release); do \
		github-release upload \
		-u ohsu-comp-bio \
		-r funnel \
		--name $$f \
		--tag dev \
		--replace \
		--file ./build/dev-release/$$f; \
	done

gce-installer: cross-compile
	mkdir -p build/gce-installer
	cp deployments/gce/bundle/* build/gce-installer/
	cp bin/funnel-linux-amd64 build/gce-installer/funnel
	cd build && \
		../deployments/gce/make-installer.sh -c gce-installer && \
		mv bundle.run funnel-gce-image-installer && \
		cd ..

gen-mocks:
	go get github.com/vektra/mockery
	mockery -dir src/funnel/scheduler/gce -name Client -print > src/funnel/scheduler/gce/mocks/Client_mock.go
	mockery -dir src/funnel/scheduler/gce -name Wrapper -print > src/funnel/scheduler/gce/mocks/Wrapper_mock.go
	mockery -dir src/funnel/server -name Database -print > src/funnel/server/mocks/Database_mock.go
	mockery -dir src/funnel/scheduler -name Database -print > src/funnel/scheduler/mocks/Database_mock.go
	mockery -dir src/funnel/scheduler -name Client -print > src/funnel/scheduler/mocks/Client_mock.go

bundle-examples:
	go-bindata -pkg examples -o src/funnel/examples/bundle.go examples

full: proto install prune_deps add_deps tidy lint test web

clean:
	rm -rf ./bin ./pkg ./test_tmp ./build ./buildtools

.PHONY: proto web

