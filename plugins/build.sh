#!/bin/bash

set -eu
set -o pipefail

mkdir -p plugin-binaries
go build -o ./plugin-binaries/exampleAuthorizer ./sample-plugins/exampleAuthorizer/
go build -o ./authorize ./main.go
