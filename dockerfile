# Build stage
FROM golang:1.24-alpine AS builder

# Install protobuf compiler
RUN apk add --no-cache protobuf protobuf-dev make

# Install protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf
RUN make proto

# Build daemon
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o vsysmon-daemon ./cmd/daemon

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /

# Copy binary from builder
COPY --from=builder /build/vsysmon-daemon .

# Copy example config
COPY --from=builder /build/config.json /etc/vsysmon/config.json

# Note: For reading /proc and /sys, container needs to run with --privileged
# or use host network mode. This is handled in docker-compose.yml

EXPOSE 50051

ENTRYPOINT ["./vsysmon-daemon"]
CMD ["-p", "50051", "-c", "/etc/vsysmon/config.json", "-v"]