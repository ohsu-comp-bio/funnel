# build stage
FROM golang:1.24-alpine AS build-env
RUN apk add --no-cache make git bash build-base
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"

WORKDIR /go/src/github.com/ohsu-comp-bio/funnel
COPY go.* .
RUN go mod download
COPY . .
# This 'make build' produces the Funnel binary in the WORKDIR
RUN --mount=type=cache,target=/root/.cache/go-build make build

# final stage for development
FROM alpine

# Install Go, Git, and essential build tools
RUN apk add --no-cache go git make bash build-base

# Set up Go environment variables
ENV GOPATH=/go
# Add Go's binary directory and common user binary locations to PATH
ENV PATH="/go/bin:/usr/local/bin:${PATH}"

# Copy the Go module cache from the build-env.
# This helps avoid re-downloading dependencies when building in the final image.
# Ensure that /go/pkg exists and has the correct permissions after copying.
COPY --from=build-env /go/pkg /go/pkg

# Set the working directory to where the source code will reside, mirroring the build environment
WORKDIR /go/src/github.com/ohsu-comp-bio/funnel

# Copy the entire source code (including go.mod, go.sum, and your project files) from the build-env.
# This makes the project ready for modification and building within the container.
COPY --from=build-env /go/src/github.com/ohsu-comp-bio/funnel /go/src/github.com/ohsu-comp-bio/funnel

# Also copy the initially built Funnel binary from the build stage.
# This can serve as a baseline or be run if no new build is performed.
# Placing it in /usr/local/bin makes it available in the PATH.
COPY --from=build-env /go/src/github.com/ohsu-comp-bio/funnel/funnel /usr/local/bin/funnel

# Expose necessary ports for Funnel
EXPOSE 8000 9090

# For a development image, this command keeps the container running,
# allowing you to 'exec' into it for development tasks (e.g., git pull, make build).
CMD ["sleep", "infinity"]