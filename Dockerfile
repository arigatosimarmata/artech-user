# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the migration CLI
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate ./cmd/migrate

# Build the HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/http

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Create app directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/migrate .
COPY --from=builder /app/server .

# Copy migration files
COPY --from=builder /app/migration ./migration

# Create logs directory
RUN mkdir -p logs

# Expose port
EXPOSE 8080

# Default command (run server)
CMD ["./server", "start"]
