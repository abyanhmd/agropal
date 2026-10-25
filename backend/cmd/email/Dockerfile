# Builder image
FROM golang:1.25-alpine AS builder
LABEL name="go-worker-email"

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/worker ./cmd/email/main.go

# Runner image
FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata \
	&& adduser -D -u 1000 appuser

COPY --from=builder /app/worker /worker

USER appuser

ENTRYPOINT ["/worker"]
