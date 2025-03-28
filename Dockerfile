# Build stage
FROM golang:1.23.4 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /oracle-backend cmd/oracle/main.go

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /oracle-backend .
COPY config.yaml .
EXPOSE 8080
CMD ["./oracle-backend"]