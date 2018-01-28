#!/bin/bash

# Must be run from the repo root.

# Kill the background processes (watching JS and SASS) on exit.
trap 'kill $(jobs -p)' EXIT

# Build the files once synchronously, so they exist before go-bindata runs below.
./webdash/node_modules/.bin/browserify webdash/app.js -d -o webdash/bundle.js
./webdash/node_modules/.bin/node-sass webdash/style.scss webdash/style.css

# Build a bundle that points to the files on the filesystem, instead of including
# the data in the generated code.
go-bindata -debug -pkg webdash -ignore node_modules -o webdash/web.go webdash build/webdash

# Build funnel with the debug dashboard bundle.
echo 'building funnel'
make install

echo 'watching'
# Watch the JS and SASS for changes and automatically rebuild.
./webdash/node_modules/.bin/watchify webdash/app.js -d -o webdash/bundle.js &
./webdash/node_modules/.bin/node-sass --watch webdash/style.scss webdash/style.css &

# block forever
cat
