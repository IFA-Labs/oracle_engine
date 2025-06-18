# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH="/go/bin:${PATH}"

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN pwd && swag --version && ls -la

RUN swag init -g internal/server/api/api.go
RUN CGO_ENABLED=0 GOOS=linux go build -o oracle-backend cmd/oracle/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/oracle-backend .
COPY config.yaml .
COPY web/ ./web/
EXPOSE 5001
EXPOSE 8000
EXPOSE 8080
CMD ["./oracle-backend"]