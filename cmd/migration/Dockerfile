# Builder image
FROM golang:1.25-alpine AS builder
LABEL name="go-migrate"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/migrate ./cmd/migration/main.go

# Runner image
FROM alpine:latest

WORKDIR /app

RUN adduser -D -u 1000 appuser

COPY --from=builder /app/migrate /migrate

USER appuser

ENTRYPOINT ["/migrate"]
CMD ["up"]
