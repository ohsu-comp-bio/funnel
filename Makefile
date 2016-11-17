GOPATH := $(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

PROTO_INC= -I ./ -I $(GOPATH)/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/

server:
	go install tes-server
	go install tes-worker

proto_build:
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
	@find ./src/tes* -type f | grep -v ".pb." | grep -E '.*\.go$$' | xargs golint
	@for d in $(GOPATH)/src/tes*; \
	do \
		go tool vet $$d; \
	done
