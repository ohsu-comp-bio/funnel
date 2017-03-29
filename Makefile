GOPATH := $(shell pwd)/build/tools:$(shell pwd)/build
export GOPATH
PATH := ${PATH}:$(shell pwd)/bin
export PATH
PYTHONPATH := ${PYTHONPATH}:$(shell pwd)/python
export PYTHONPATH

PROTO_INC=-I ./ -I $(shell pwd)/funnel/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis
GRPC_HTTP_MOD=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api

server: depends
	go install funnel

depends:
	mkdir -p build/src build/bin build/pkg build/tools
	git submodule update --init --recursive
	ln -s $(shell pwd)/funnel/ $(shell pwd)/build/src/funnel
	go get -d funnel

proto:
	@go get github.com/golang/protobuf/protoc-gen-go/
	@go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/
	@cd ./funnel/proto/tes && protoc \
		$(PROTO_INC) \
		--go_out=$(GRPC_HTTP_MOD),plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		tes.proto
	@cd ./funnel/proto/funnel && protoc \
		$(PROTO_INC) \
		-I ../tes \
		--go_out=$(GRPC_HTTP_MOD),Mtes.proto=funnel/proto/tes,plugins=grpc:. \
		--grpc-gateway_out=logtostderr=true:. \
		funnel.proto
	@find ./funnel/proto -name *pb* -type f -exec sed -i '' 's/ga4gh_task_exec/tes/g' {} +

serve-doc:
	godoc --http=:6060

add_deps:
	go get github.com/dpw/vendetta
	./build/tools/bin/vendetta ./funnel

prune_deps:
	go get github.com/dpw/vendetta
	./build/tools/bin/vendetta -p ./funnel

tidy:
	pip install -q autopep8
	@find ./funnel -type f | grep -v "funnel/vendor" | grep -v ".pb." | grep -E '.*\.go$$' | xargs gofmt -w -s
	@find ./* -type f | grep -E '.*\.py$$' | grep -v "/venv/" | grep -v "/web/node" | xargs autopep8 --in-place --aggressive --aggressive

lint:
	pip install -q flake8
	flake8 --exclude ./venv,./web .
	go get github.com/alecthomas/gometalinter
	./build/tools/bin/gometalinter --install > /dev/null
	./build/tools/bin/gometalinter --disable-all --enable=vet --enable=golint --enable=gofmt --vendor -s proto ./funnel/...

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
