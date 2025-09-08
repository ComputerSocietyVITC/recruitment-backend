# Build stage
FROM golang:1.24.6-alpine AS builder

# Install git (required for some Go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration files and wait script
COPY --from=builder /app/models/migrations ./models/migrations/
COPY --from=builder /app/wait-for-postgres.sh .

# Make the wait script executable and change ownership to non-root user
RUN chmod +x wait-for-postgres.sh && \
    chown -R appuser:appgroup /root

# Install postgresql-client for psql command
USER root
RUN apk add --no-cache postgresql-client

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]
