.PHONY: all build

all: build

build:
	@printf "Building ./server..."
	@mkdir -p plugin-binaries
	@go build -o ./plugin-binaries/exampleAuthorizer ./authorizer/
	@go build -o ./server ./main.go
	@echo "OK"
	