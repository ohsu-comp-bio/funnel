#!/bin/bash

export GIT_COMMIT=$(git rev-parse --short HEAD)
export GIT_BRANCH=$(git symbolic-ref -q --short HEAD)
export GIT_REF=$(git symbolic-ref -q --short HEAD)
export GIT_URL=$(git config branch.$GIT_REF.remote)
export GIT_UPSTREAM=$(git remote get-url $GIT_URL 2> /dev/null)

./build/goreleaser-linux --rm-dist --snapshot
