# Stage 1: Build
FROM golang:1.25-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && go mod download && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o test-app-backend ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache wget
RUN adduser -D -g '' appuser
WORKDIR /app
COPY --from=builder /app/test-app-backend .
USER appuser
EXPOSE 8081
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8081/health || exit 1
CMD ["./test-app-backend"]
