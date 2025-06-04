# build stage
FROM golang:1.24-alpine AS build-env
RUN apk add --no-cache make git bash build-base
ENV GOPATH=/go
ENV PATH="/go/bin:${PATH}"

WORKDIR /go/src/github.com/ohsu-comp-bio/funnel
COPY go.* .
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build make build

# deployment stage
FROM alpine
RUN apk add --no-cache go git make bash build-base

ENV GOPATH=/go
ENV PATH="/app:${PATH}"

WORKDIR /funnel
COPY --from=build-env /go/pkg /go/pkg

# Funnel source code
COPY --from=build-env /go/src/github.com/ohsu-comp-bio/funnel /funnel

# Funnel executable
COPY --from=build-env  /go/src/github.com/ohsu-comp-bio/funnel/funnel /app/funnel
EXPOSE 8000 9090

ENTRYPOINT ["/app/funnel"]
