# Agropal

`backend/` is the Go API, `frontend/` is the Next.js app. This folder runs both together.

## Run everything

```sh
cp .env.example .env      # then change the passwords and secrets
docker compose up --build
```

| What | Where |
| --- | --- |
| App | http://localhost:3000 |
| API | http://localhost:8080/api/v1 |
| Sign-in emails (mail catcher) | http://localhost:8025 |

Signing in sends a one-time code by email. With the default setup nothing is really sent: open the mail
catcher and copy the code from there.

Stop with `docker compose down`; add `-v` to also delete the database.

## Things to know

- The API allows browser calls from the origins in `backend/config/config.yaml` (`cors.allowed_origins`,
  `http://localhost:3000` by default). Change it if you serve the app from another address.
- `NEXT_PUBLIC_*` settings are baked into the frontend image when it is built, so rebuild after changing them.
- The sign-in cookie is `Secure`. Browsers accept that on `localhost`; anywhere else the app needs HTTPS, and
  the app and API must share a registrable domain (for example `app.example.com` and `api.example.com`).
- For real email, set `SMTP_*` and `SECRET_SMTP` in `.env` and remove the `mailpit` service.

## Deploy to a server

`docker-compose.prod.yaml` runs the same stack behind Caddy, which serves the app and the API from one HTTPS
origin (`/` is the frontend, `/api` is the API) and renews its certificate by itself. Only ports 80 and 443 are
published, and there is no mail catcher, so a real SMTP provider is required.

```sh
cp .env.prod.example .env      # set DOMAIN, passwords, secrets (openssl rand -hex 32) and SMTP_*
docker compose -f docker-compose.prod.yaml up -d --build
```

Point the domain's DNS A record at the server first. About 2 vCPU and 4 GB RAM is comfortable. The API
reports liveness at `/healthz` and refuses to start with a JWT or GCM secret under 32 characters.
