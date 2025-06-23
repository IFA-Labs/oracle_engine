# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install swag and other dependencies
RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH="/go/bin:${PATH}"

# Install system dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Debug: Show current directory and files
RUN pwd && ls -la

# Generate swagger docs
RUN swag init -g internal/server/api/api.go --output docs

# Debug: Check if swagger files were generated
RUN ls -la docs/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o oracle-backend cmd/oracle/main.go

# Debug: Check if binary was created and show its details
RUN ls -la oracle-backend && file oracle-backend

# Run stage
FROM alpine:latest
WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder stage
COPY --from=builder /app/oracle-backend .

# Copy configuration files
COPY config.yaml .
COPY web/ ./web/

# Make sure the binary is executable
RUN chmod +x oracle-backend

# Debug: Check if binary exists and is executable
RUN ls -la oracle-backend && file oracle-backend

EXPOSE 5001
EXPOSE 8000
EXPOSE 8080

CMD ["./oracle-backend"]