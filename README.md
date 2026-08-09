# Agropal

Agropal is a backend platform for managing farms and farming activity: farms and their weather locations, products with price/stock history, transactions, social posts/comments, and user accounts, all behind a JWT-authenticated REST API.

## Tech Stack

- **Language:** Go 1.24
- **Database:** PostgreSQL (`pgx/v5`, `lib/pq`, `sqlx`)
- **Cache/Session:** Redis (configured, not yet used by any service)
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
│   ├── email/                # (scaffolded, empty)
│   └── migration/            # (scaffolded, empty; migrations currently run manually)
├── config/                   # config.yaml + env-backed Config loader
├── internal/
│   ├── api/                  # HTTP layer, one package per domain
│   │   ├── auth/             #   register/login/refresh/logout
│   │   ├── farm/             #   farm CRUD
│   │   ├── weather/          #   weather locations per farm
│   │   ├── product/          #   products + price/stock history
│   │   ├── transaction/      #   income/expense transactions
│   │   ├── post/             #   posts, comments, like/dislike/view engagement
│   │   ├── user/             #   profile, password change, admin user listing
│   │   └── middleware/       #   JWT auth middleware + route registration helper
│   ├── app/                  # use cases, one package per domain + container.go
│   ├── domain/                # entities, repository interfaces, error keys
│   │   ├── auth/ farm/ note/ post/ product/ transaction/ user/ weather/
│   │   └── shared/           #   context/claims helpers, common error & success keys
│   ├── infra/
│   │   ├── crypto/           #   JWT signing/validation, AES-GCM encryption
│   │   ├── database/         #   Postgres connection + transactor
│   │   ├── identity/         #   ID/token generators
│   │   └── persistence/      #   sqlx-based repository implementations
│   ├── migration/            # raw SQL migrations (up/down pairs)
│   ├── request/               # HTTP request DTOs + ToCommand() mappers
│   ├── response/              # HTTP response DTOs + error/success registries
│   └── utils/                 # IOHandler (JSON I/O + response envelopes), validators
├── pkg/                      # shared infra-agnostic packages (logger)
└── logs/                     # log output
```

## API Surface

All routes are mounted under `/api/v1`. Everything except `/auth/*` requires `Authorization: Bearer <access token>`.

| Domain | Routes |
|---|---|
| Auth | `POST /auth/register`, `/login`, `/refresh`, `/logout` |
| Farm | `POST/GET /farm`, `GET/PATCH/DELETE /farm/{farm_id}` |
| Weather | `POST/GET /farm/{farm_id}/weather`, `PATCH/DELETE /farm/{farm_id}/weather/{weather_id}` |
| Product | `POST/GET /farm/{farm_id}/product`, `GET/PATCH/DELETE .../product/{product_id}`, plus nested `/price` and `/stock` create+delete |
| Transaction | `POST/GET/DELETE .../product/{product_id}/transaction`, `GET/PATCH .../transaction/{transaction_id}`, `GET /farm/{farm_id}/transaction`, bulk `DELETE /transaction` and `/transaction/all` |
| Post | `POST/GET /post`, `GET /post/{post_id}`, `POST /post/{post_id}/comment`, `POST /post/{post_id}/engagement`, `POST /comment/{comment_id}/engagement` |
| User | `GET/PATCH/DELETE /user/{user_id}` (or `me`), `PATCH /user/{user_id}/password`, `GET /users` |

## Response Conventions

Mutating endpoints (`POST` create, `PATCH`, `DELETE`) respond with a success envelope (`{info: {code, key, message}, data, request_id}`) instead of a bare status with no body. `key`/`message` are looked up in `response.SuccessRegistry` (`internal/response/registry_success.go`) by the domain's `SuccessKey` (e.g. `dFarm.SuccessCreated` → `"FARM.CREATE_SUCCESS"`), defined alongside each domain's error keys in `internal/domain/<domain>/entity.go`. `GET` endpoints and `POST /auth/login`/`/auth/refresh` return the raw DTO instead, and errors use the equivalent `ErrorRegistry`-backed envelope. See `docs/openapi.yaml` and `docs/README.md` for the full endpoint-by-endpoint mapping.

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL
- Redis (reachable, though nothing depends on it functioning yet)

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

SECRET_JWT=
SECRET_GCM_TOKEN=
SECRET_SMTP=
```

Server and path settings (host, port, migration/log paths) live in `backend/config/config.yaml`.

### Database

Apply the SQL files under `internal/migration/` (numbered `up`/`down` pairs) against your Postgres instance in order; there's no migration runner wired up yet (`cmd/migration` is a placeholder).

### Running

```bash
cd backend
go run ./cmd/api
```

The server listens on the host/port from `config/config.yaml` (default `localhost:8080`).

### Build & verify

```bash
go build ./...
go vet ./...
gofmt -l .
```

There is no automated test suite yet.

## Domains

- **Auth**: registration, login, refresh-token rotation (AES-GCM encrypted at rest), logout, password reset (token issued via the refresh-token store)
- **User**: profile (name/avatar), password change, soft/hard delete, listing users by related asset
- **Farm**: a user's farms
- **Weather**: weather-tracked locations (lat/long) attached to a farm
- **Product**: a farm's products, with append-only price and stock history
- **Transaction**: income/expense records against a product
- **Post**: posts with comments and like/dislike/view engagement, backed by an audit trail in `domain/note`

## Known Limitations

- Not every mutation has a dedicated success key yet; post/comment engagement endpoints (`POST /post/{post_id}/engagement`, `POST /comment/{comment_id}/engagement`) still return a bare `201` with no body.
- Struct `validate` tags exist on several commands but no validation library is wired in yet; they're currently documentation only.
- `app/product`'s `UpdateProduct` records an audit note for changed fields but doesn't yet persist the new values.
- `cmd/email`, `cmd/migration`, and `boilerplate/*` are scaffolded but empty.
- No automated tests.
