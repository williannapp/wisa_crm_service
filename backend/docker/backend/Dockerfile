# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o wisa-crm-service ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache wget
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/wisa-crm-service .
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1
CMD ["./wisa-crm-service"]
