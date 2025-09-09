# =============================================================================
# Multi-stage Dockerfile for Go recruitment backend
# =============================================================================

# Build arguments for flexibility
ARG GO_VERSION=1.25.1
ARG ALPINE_VERSION=3.22

# Stage 1: Build stage using official Go image
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Install ca-certificates for HTTPS requests, git for go modules, and upx for binary compression
RUN apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    upx

# Create app user for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod and sum files for dependency caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum haven't changed)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" removes debug info and symbol table to reduce size
# -trimpath removes absolute paths from compiled executable
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -a -installsuffix cgo \
    -o recruitment-backend \
    ./main.go

# Compress binary with UPX (optional, can reduce size by ~50-70%)
# Comment out if you prefer faster startup over smaller size
RUN upx --best --lzma recruitment-backend

# =============================================================================
# Stage 2: Final runtime stage using minimal distroless image
FROM gcr.io/distroless/static-debian12:nonroot

# Copy ca-certificates and timezone data from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compressed binary from builder stage
COPY --from=builder /app/recruitment-backend /recruitment-backend

# Copy essential application files
COPY --from=builder /app/models/migrations ./models/migrations
COPY --from=builder /app/utils ./utils

# Use non-root user from distroless (uid: 65532, gid: 65532)
USER nonroot:nonroot

# Expose the port the app runs on
EXPOSE 8080

# Set environment variables for production
ENV ENV=production
ENV GIN_MODE=release
ENV PORT=8080
ENV TZ=UTC

# Health check using the application itself
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/recruitment-backend", "--health-check"]

# Run the binary
ENTRYPOINT ["/recruitment-backend"]

# =============================================================================
# Alternative minimal scratch-based stage (even smaller but less features)
# =============================================================================
FROM scratch AS minimal

# Copy ca-certificates from builder for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage
COPY --from=builder /app/recruitment-backend /recruitment-backend

# Copy essential application files
COPY --from=builder /app/models/migrations ./models/migrations
COPY --from=builder /app/utils ./utils

# Expose the port
EXPOSE 8080

# Set environment variables
ENV ENV=production
ENV GIN_MODE=release
ENV PORT=8080

# Run the binary
ENTRYPOINT ["/recruitment-backend"]
