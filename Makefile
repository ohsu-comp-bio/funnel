
GOPATH := $(shell pwd)
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH

PROTO_INC= -I ./ -I $(GOPATH)/src/github.com/gengo/grpc-gateway/third_party/googleapis/

grpc : 
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/gengo/grpc-gateway/protoc-gen-swagger
	go get -u github.com/gengo/grpc-gateway/protoc-gen-grpc-gateway
	cd task-execution-schemas/proto && protoc $(PROTO_INC) \
		--go_out=Mgoogle/api/annotations.proto=github.com/gengo/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../../src/ga4gh-tasks/ \
		--grpc-gateway_out=logtostderr=true:../../src/ga4gh-tasks/ \
		task_execution.proto
