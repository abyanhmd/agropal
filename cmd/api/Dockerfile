# Builder image
FROM golang:1.25-alpine AS builder
LABEL name="go-api"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api/main.go

# Runner image
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache curl net-tools tzdata \
	&& adduser -D -u 1000 appuser

COPY --from=builder /app/api /api

USER appuser

ENTRYPOINT ["/api"]
