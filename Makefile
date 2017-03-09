GOPATH := $(shell pwd)/buildtools:$(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

PROTO_INC= -I ./ -I $(GOPATH)/src/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/
GRPC_HTTP_MOD=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api

server: depends
	go install tes-server
	go install tes-worker

proto_build:
	@go get ./src/vendor/github.com/golang/protobuf/protoc-gen-go/
	@go get ./src/vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/
	@cd task-execution-schemas/proto && protoc $(PROTO_INC) \
		--go_out=$(GRPC_HTTP_MOD),plugins=grpc:../../src/tes/ga4gh/ \
		--grpc-gateway_out=logtostderr=true:../../src/tes/ga4gh/ \
		task_execution.proto
	@cd proto && protoc \
		$(PROTO_INC) \
		-I ../task-execution-schemas/proto/ \
		--go_out=$(GRPC_HTTP_MOD),Mtask_execution.proto=tes/ga4gh,plugins=grpc:../src/tes/server/proto \
		--grpc-gateway_out=logtostderr=true:../src/tes/server/proto/ \
		task_worker.proto

depends:
	git submodule update --init --recursive
	go get -d tes-server
	go get -d tes-worker

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
	@find ./src/tes* -type f | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w -s
	@find ./* -type f | grep -E '.*\.py$$' | grep -v "/venv/" | grep -v "/share/node" | xargs autopep8 --in-place --aggressive --aggressive

lint:
	flake8 --exclude ./venv .
	go get github.com/alecthomas/gometalinter
	./buildtools/bin/gometalinter --install > /dev/null
	./buildtools/bin/gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --vendor -s ga4gh -s proto ./src/...

test:	
	docker build -t tes-wait -f tests/docker_files/tes-wait/Dockerfile tests/docker_files/tes-wait/
	pip2.7 install -q -r tests/requirements.txt
	nosetests-2.7 tests/

web:
	cd share && \
	npm install && \
	./node_modules/node-sass/bin/node-sass style.scss style.css && \
	cd ..
