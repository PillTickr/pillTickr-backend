# Stage 1: Build
FROM golang:1.24-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Install git (if needed for module imports)
RUN apk add --no-cache git

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . .

# Build the binary
RUN go build -o pilltickr main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Copy binary and .env
COPY --from=builder /app/pilltickr .
COPY stack.env ./.env

EXPOSE 8080

ENTRYPOINT ["./pilltickr"]
