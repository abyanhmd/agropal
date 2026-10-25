# Builder image
FROM golang:1.25-alpine AS builder
LABEL name="go-api"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api/main.go

# Runner image
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache tzdata \
	&& adduser -D -u 1000 appuser

COPY --from=builder /app/api /api

USER appuser

# busybox wget ships with alpine
HEALTHCHECK --interval=15s --timeout=3s --start-period=20s --retries=3 \
	CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/api"]
