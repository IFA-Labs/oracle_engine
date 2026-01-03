# ==============================================================================
# Build stage
# ==============================================================================
FROM golang:1.25-alpine AS builder

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Install swag for swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH="/go/bin:${PATH}"

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Generate swagger docs
RUN swag init -g internal/server/api/api.go --output docs

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o oracle-backend \
    cmd/oracle/main.go

# ==============================================================================
# Production stage
# ==============================================================================
FROM alpine:3.19

# Labels for container metadata
LABEL org.opencontainers.image.title="Oracle Engine" \
    org.opencontainers.image.description="IFA Labs Oracle Engine - Price Feed Service" \
    org.opencontainers.image.vendor="IFA Labs" \
    org.opencontainers.image.source="https://github.com/IFA-Labs/oracle_engine"

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl && \
    # Create non-root user for security
    addgroup -g 1001 -S oracle && \
    adduser -u 1001 -S oracle -G oracle && \
    # Create necessary directories
    mkdir -p /app/logs && \
    chown -R oracle:oracle /app

# Copy the binary from builder stage
COPY --from=builder --chown=oracle:oracle /app/oracle-backend .

# Copy configuration files
COPY --chown=oracle:oracle config.yaml .
COPY --chown=oracle:oracle web/ ./web/

# Switch to non-root user
USER oracle

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Expose ports
EXPOSE 8000

# Run the application
ENTRYPOINT ["./oracle-backend"]