GOPATH := $(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

PROTO_INC= -I ./ -I $(GOPATH)/src/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/

server: depends
	go install tes-server
	go install tes-worker

proto_build:
	go get ./src/vendor/github.com/golang/protobuf/protoc-gen-go/
	go get ./src/vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/

	cd task-execution-schemas/proto && protoc $(PROTO_INC) \
		--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../../src/tes/ga4gh/ \
		--grpc-gateway_out=logtostderr=true:../../src/tes/ga4gh/ \
		task_execution.proto
	cd proto && protoc \
		$(PROTO_INC) \
		-I ../task-execution-schemas/proto/ \
		--go_out=Mtask_execution.proto=tes/ga4gh,plugins=grpc:../src/tes/server/proto \
		task_worker.proto

depends:
	go get -d tes-server/
	go get -d tes-worker/

golint:
	go get -v github.com/golang/lint/golint/

tidy: golint
	@find ./src/tes* -type f | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w
	@for d in `find ./src/tes -type d | grep -E -v "ga4gh|proto"`; do \
		echo $$d; \
		golint $$d; \
	done
	@for d in $(GOPATH)/src/tes*; do \
		go tool vet $$d; \
	done
