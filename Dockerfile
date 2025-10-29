FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o godfs ./cmd/master/main.go
RUN go build -o chunkserver ./cmd/chunkserver/main.go
RUN go build -o web ./cmd/web/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binaries
COPY --from=builder /app/godfs .
COPY --from=builder /app/chunkserver .
COPY --from=builder /app/web .

# Copy source for go run commands
COPY --from=builder /app .

# Create data directories
RUN mkdir -p chunkserver_data_1 chunkserver_data_2 chunkserver_data_3

EXPOSE 9000 9001 9002 9003 8080

CMD ["./godfs"]
