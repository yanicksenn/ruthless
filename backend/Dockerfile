# Build stage
FROM golang:1.25.0-alpine AS build

# Install dependencies for protoc
RUN apk add --no-cache protobuf-dev protoc

# Install protoc plugins for Go
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Generate Proto code
RUN mkdir -p api/v1/ruthlespb
RUN protoc --proto_path=. \
    --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/v1/ruthless.proto api/v1/config.proto

# Build the backend binary
RUN go build -o /cah github.com/yanicksenn/ruthless/backend/cmd/cah

# Final stage
FROM alpine:latest

# Install CA certificates for TLS (if needed)
RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /cah /usr/local/bin/cah
COPY --from=build /app/backend/config.textproto /app/backend/config.textproto

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/cah"]
