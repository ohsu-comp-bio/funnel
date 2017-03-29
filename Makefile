GOPATH := $(shell pwd)/buildtools:$(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH
PYTHONPATH := ${PYTHONPATH}:$(shell pwd)/python
export PYTHONPATH

PROTO_INC=-I ./ -I $(GOPATH)/src/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
GRPC_HTTP_MOD=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api

server: depends
	go install funnel

proto_build:
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
	@find src/funnel/proto -name *pb* -type f -exec sed -i '' 's/ga4gh_task_exec/tes/g' {} +

depends:
	git submodule update --init --recursive
	go get -d funnel

serve-doc:
	godoc --http=:6060

add_deps:
	go get github.com/dpw/vendetta
	./buildtools/bin/vendetta src/

prune_deps:
	go get github.com/dpw/vendetta
	./buildtools/bin/vendetta -p src/

tidy:
	pip install -q autopep8
	@find ./src/funnel* -type f | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w -s
	@find ./* -type f | grep -E '.*\.py$$' | grep -v "/venv/" | grep -v "/web/node" | xargs autopep8 --in-place --aggressive --aggressive

lint:
	pip install -q flake8
	flake8 --exclude ./venv,./web .
	go get github.com/alecthomas/gometalinter
	./buildtools/bin/gometalinter --install > /dev/null
	./buildtools/bin/gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --vendor -s ga4gh -s proto ./src/funnel/...

test:	
	docker build -t tes-wait -f tests/docker_files/tes-wait/Dockerfile tests/docker_files/tes-wait/
	pip2.7 install -q -r tests/requirements.txt
	nosetests-2.7 tests/
	go test funnel/...

web:
	cd web && \
	npm install && \
	./node_modules/.bin/browserify app.js -o bundle.js && \
	./node_modules/node-sass/bin/node-sass style.scss style.css && \
	cd ..
