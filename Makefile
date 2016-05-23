
GOPATH := $(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

PROTO_INC= -I ./ -I $(GOPATH)/src/github.com/gengo/grpc-gateway/third_party/googleapis/

server:
	go install ga4gh-taskserver
	go install ga4gh-worker

proto_build: 
	cd task-execution-schemas/proto && protoc $(PROTO_INC) \
		--go_out=Mgoogle/api/annotations.proto=github.com/gengo/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../../src/ga4gh-tasks/ \
		--grpc-gateway_out=logtostderr=true:../../src/ga4gh-tasks/ \
		task_execution.proto
	cd proto && protoc \
	  $(PROTO_INC) \
	  -I ../task-execution-schemas/proto/ \
	  --go_out=Mtask_execution.proto=ga4gh-tasks,plugins=grpc:../src/ga4gh-server/proto \
		task_worker.proto
	
grpc:
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gengo/grpc-gateway/protoc-gen-swagger
	go get -u github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway
