# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o done .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 -S appuser && \
    adduser -u 1000 -S appuser -G appuser

# Create directory for database with proper permissions
RUN mkdir -p /data && chown appuser:appuser /data

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/done .

# Copy frontend files
COPY --from=builder /app/frontend ./frontend

# Change ownership of app directory
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 3001

# Volume for database persistence
VOLUME ["/data"]

# Run the application with custom db path
CMD ["./done", "-port", "3001", "-dbpath", "/data/tasks.db"]
