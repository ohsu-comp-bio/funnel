# Use the official Golang image as the base image
FROM golang:1.23-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

RUN mkdir -p ./build

# Build the Go app (CLI)
RUN go build -o ./build/test-server .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./build/cli"]
