# Dockerfile

# Stage 1: Builder
FROM golang:1.23 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Air for live reloading
RUN go install github.com/air-verse/air@1.61.1

# Copy the source code into the container
COPY . .

# Build the Go app (for production)
RUN CGO_ENABLED=0 GOOS=linux go build -o tmp/main cmd/rest-api/main.go

# Stage 2: Development Environment
FROM golang:1.23 AS dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

COPY .air.toml .air.toml

# Install Air for live reloading
RUN go install github.com/air-verse/air@v1.61.1

# Command to run the application in development mode
CMD ["air"]

# Stage 3: Production Environment
FROM alpine:latest AS production

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the builder stage
COPY --from=builder /app/main .

# Install necessary dependencies
RUN apk add --no-cache ca-certificates

# Command to run the executable in production
CMD ["./main"]
