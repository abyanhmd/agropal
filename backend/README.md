# Agropal

Agropal is a backend platform for managing farms and farming activity: farms and their weather locations, products with price/stock history, transactions, social posts/comments, and user accounts, all behind a JWT-authenticated REST API.

## Tech Stack

- **Language:** Go 1.25
- **Database:** PostgreSQL (`pgx/v5`, `lib/pq`, `sqlx`)
- **Cache/Session:** Redis (holds staged registrations, one-time verification codes, and resend cooldowns for the auth flow)
- **Auth:** JWT (`golang-jwt/jwt/v5`), AES-GCM-encrypted refresh tokens at rest
- **Config:** Viper (`spf13/viper`) + `.env` (`godotenv`)
- **Logging:** `rs/zerolog`
- **Routing:** standard library `net/http` (Go 1.22+ method-pattern mux), no external web framework

## Architecture

The backend follows a layered design, each layer depending only on the one beneath it:

```
internal/api      → HTTP handlers, request/response mapping, routing, JWT middleware
internal/app      → use cases / application services, orchestrates domain + infra
internal/domain   → entities, repository interfaces, domain errors (no framework deps)
internal/infra    → concrete implementations: Postgres persistence, crypto, identity
```

`internal/request` and `internal/response` hold the HTTP-facing DTOs shared across `internal/api` handlers, kept separate from domain entities so the wire format can evolve independently of the data model.

Within `internal/app`, `container.go` wires every service's dependencies (repositories, the transactor, the JWT provider) into a single `Container`, built once in `cmd/api/main.go` and handed to the HTTP layer.

## Project Structure

```
backend/
├── cmd/
│   ├── api/                  # HTTP server entry point (main.go)
│   ├── email/                # email worker: consumes queue.QueueNameEmail, sends via SMTP
│   └── migration/            # goose migration runner entry point (main.go)
├── config/                   # config.yaml + env-backed Config loader
├── internal/
│   ├── api/                  # HTTP layer, one package per domain
│   │   ├── auth/             #   register/login (email-code confirmed), refresh, logout, password reset
│   │   ├── farm/             #   farm CRUD
│   │   ├── location/         #   farm locations (lat/long), served under /farm/{farm_id}/location
│   │   ├── product/          #   products + price/stock history
│   │   ├── transaction/      #   income/expense transactions
│   │   ├── post/             #   posts, comments, like/dislike/view engagement
│   │   ├── user/             #   profile, password change, admin user listing
│   │   └── middleware/       #   JWT auth middleware + route registration helper
│   ├── app/                  # use cases, one package per domain + container.go
│   ├── domain/                # entities, repository interfaces, error keys
│   │   ├── auth/ farm/ note/ post/ product/ transaction/ user/ location/
│   │   └── shared/           #   context/claims helpers, common error & success keys
│   ├── infra/
│   │   ├── crypto/           #   JWT signing/validation, AES-GCM encryption
│   │   ├── database/         #   Postgres connection + transactor
│   │   ├── email/             #   SMTP sender + verification email templates
│   │   ├── identity/         #   ID/token generators
│   │   ├── queue/            #   RabbitMQ publish/consume
│   │   └── persistence/      #   sqlx-based repository implementations
│   ├── migration/            # goose SQL migrations (one file per version, Up+Down)
│   ├── request/               # HTTP request DTOs + ToCommand() mappers
│   ├── response/              # HTTP response DTOs + error/success registries
│   └── utils/                 # IOHandler (JSON I/O + response envelopes), validators
├── pkg/                      # shared infra-agnostic packages (logger)
└── logs/                     # log output
```

## API Surface

All routes are mounted under `/api/v1`. Everything except `/auth/*` requires `Authorization: Bearer <access token>`.

Auth uses a two-step email flow. `POST /auth/register` and `POST /auth/login` only email a one-time code; the matching `/verify` call confirms the code and returns the token pair. `POST /auth/{purpose}/resend-code` re-sends a code for `purpose` `register` or `reset_password`. Password reset is `reset-password/request` followed by `reset-password/confirm`.

| Domain | Routes |
|---|---|
| Auth | `POST /auth/register`, `/register/verify`, `/login`, `/login/verify`, `/{purpose}/resend-code`, `/refresh`, `/logout`, `/reset-password/request`, `/reset-password/confirm` |
| Farm | `POST/GET /farm`, `GET/PATCH/DELETE /farm/{farm_id}` |
| Location | `POST/GET /farm/{farm_id}/location`, `PATCH/DELETE /farm/{farm_id}/location/{location_id}` |
| Product | `POST/GET /farm/{farm_id}/product`, `GET/PATCH/DELETE .../product/{product_id}`, plus nested `/price` and `/stock` create+delete |
| Transaction | `POST/GET/DELETE .../product/{product_id}/transaction`, `GET/PATCH .../transaction/{transaction_id}`, `GET /farm/{farm_id}/transaction`, bulk `DELETE /transaction` and `/transaction/all` |
| Post | `POST/GET /post`, `GET /post/{post_id}`, `POST /post/{post_id}/comment`, `POST/DELETE /post/{post_id}/engagement`, `POST/DELETE /comment/{comment_id}/engagement` |
| User | `GET/PATCH/DELETE /user/{user_id}` (or `me`), `PATCH /user/{user_id}/password`, `GET /users` |

## Response Conventions

Mutating endpoints (`POST` create, `PATCH`, `DELETE`) respond with a success envelope (`{info: {code, key, message}, data, request_id}`) instead of a bare status with no body. `key`/`message` are looked up in `response.SuccessRegistry` (`internal/response/registry_success.go`) by the domain's `SuccessKey` (e.g. `dFarm.SuccessCreated` → `"FARM.CREATE_SUCCESS"`), defined alongside each domain's error keys in `internal/domain/<domain>/entity.go`. `GET` endpoints return the raw DTO instead, and errors use the equivalent `ErrorRegistry`-backed envelope. The `/auth/*` endpoints also use the success envelope; when a token pair is issued (the `/verify` endpoints and `/auth/refresh`) it sits under the envelope's `data` field. See `docs/openapi.yaml` and `docs/README.md` for the full endpoint-by-endpoint mapping.

Every request is assigned a server-generated correlation id (a UUID) by the `RequestID` middleware; any inbound `X-Request-Id` is ignored, so the backend is the sole source of truth. It is echoed on the response's `X-Request-Id` header, included as `request_id` in every envelope (success and error), and attached to the matching access-log line, so a client-reported id can be traced to a server-side log entry.

Request bodies are validated with [go-playground/validator](https://github.com/go-playground/validator) via `validate` struct tags on the `internal/request` DTOs (`internal/utils/validate.go`). A handler calls `ioHandler.ValidateStruct(ctx, w, req)` right after decoding the body; on failure it writes the `COMMON.VALIDATION_FAILED` error with the envelope's `errors` field populated as `{field: [messages]}`, one entry per failing field. Endpoints with validation semantics that don't map cleanly onto struct tags (e.g. auth's email/password rules, which need specific per-rule error keys) keep their own hand-written checks in `internal/api/<domain>/mapper.go` instead.

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis (the auth flow stores staged registrations and verification codes here, so the server needs it running)
- RabbitMQ (the api publishes verification emails to it; run `cmd/email` to consume and send them)

### Configuration

Create a `.env` file in `backend/` with the following keys:

```
POSTGRES_HOST=
POSTGRES_PORT=
POSTGRES_USER=
POSTGRES_PASSWORD=
POSTGRES_NAME=

REDIS_ADDRESS=
REDIS_PASSWORD=
REDIS_DB=
REDIS_PROTOCOL=

SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_FROM=

SECRET_JWT=
SECRET_GCM_TOKEN=
SECRET_SMTP=
```

`SECRET_SMTP` is the SMTP account's password/app-password; `SMTP_USERNAME`/`SMTP_FROM` are usually the same address.

Server and path settings (host, port, migration/log paths) live in `backend/config/config.yaml`.

If the server runs behind a reverse proxy (the docker-compose stack puts Traefik in front of it), set `server.trusted_proxies` in `config.yaml` to the proxy's IP ranges, for example `["172.16.0.0/12"]` for the compose bridge network. Rate limiting and request logging then read the real client IP from `X-Forwarded-For` instead of keying every request to the proxy. Left empty, the connection peer address is used as-is.

For browser (website) clients, list the front-end origins in `cors.allowed_origins` in `config.yaml`, for example `["https://app.example.com", "http://localhost:5173"]`. The `middleware.CORS` wrapper then answers preflight `OPTIONS` and sends `Access-Control-Allow-Origin`/`-Credentials` for those origins (credentials are on, since the auth flow uses the `rid` cookie), and exposes `X-Request-Id` to page scripts. The single entry `"*"` reflects any origin — development only. Native/mobile and server-to-server clients send no `Origin` header and are unaffected either way; an empty list simply blocks browser cross-origin calls.

### Database

Migrations are managed with [goose](https://github.com/pressly/goose) via `cmd/migration`. Each file under `internal/migration/` contains both the `-- +goose Up` and `-- +goose Down` sections for one migration.

```bash
go run ./cmd/migration up       # apply all pending migrations
go run ./cmd/migration down     # roll back the most recent migration
go run ./cmd/migration status   # show applied/pending migrations
go run ./cmd/migration redo     # roll back and reapply the last migration
```

To add a new migration, create a file like `internal/migration/00008_<name>.sql` with `-- +goose Up` / `-- +goose Down` sections, or generate the skeleton with the goose CLI (`goose -dir internal/migration create <name> sql`).

### Running

```bash
cd backend
go run ./cmd/api
```

The server listens on the host/port from `config/config.yaml` (`0.0.0.0:8080`, reachable at `localhost:8080` locally).

Run the email worker (requires RabbitMQ and the `SMTP_*`/`SECRET_SMTP` env vars) alongside it to actually deliver verification codes:

```bash
go run ./cmd/email
```

### Running with Docker

```bash
docker compose -f docker-compose-local.yaml up --build
```

This builds and runs `go-api`, `go-migrate` (applies migrations once, then exits), `go-worker-email`, plus Postgres/Redis/RabbitMQ and a Traefik reverse proxy, all wired together on an internal Docker network. The api is reachable at `http://api.localhost` (routed through Traefik on port 80), RabbitMQ's management UI at `http://rabbitmq.localhost` or `localhost:15672`.

`.env`'s `POSTGRES_HOST`/`REDIS_ADDRESS`/`RABBITMQ_ADDRESS` are written for running the binaries directly on the host (`go run`); the compose file overrides those three per-service to the in-network container names (`postgres`, `redis`, `rabbitmq`) so the containerized services can actually reach each other. If you add a new inter-service dependency, remember to add the matching override rather than relying on `.env`'s host values.

### Build & verify

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

The same steps run in CI on every push and pull request against `master` (`.github/workflows/ci.yml`), with tests run under `-race`.

## Domains

- **Auth**: registration and login (both confirmed by a one-time code emailed to the user), refresh-token rotation (AES-GCM encrypted at rest), logout, password reset by emailed code
- **User**: profile (name/avatar), password change, soft/hard delete, listing users by related asset
- **Farm**: a user's farms
- **Location**: weather-tracked locations (lat/long) attached to a farm
- **Product**: a farm's products, with append-only price and stock history
- **Transaction**: income/expense records against a product
- **Post**: posts with comments and like/dislike/view engagement, backed by an audit trail in `domain/note`

## Known Limitations

- `boilerplate/*` is scaffolded but empty.
- `domain/note` is a write-only audit trail by design: users, products and posts record notes as they change, and nothing reads them back through the API. Add `app/note` and `api/note` if that is ever wanted.
