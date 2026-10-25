# Agropal API: OpenAPI Spec

`openapi.yaml` is an OpenAPI 3.0.3 spec covering all HTTP routes exposed by `cmd/api`.

## Import into Postman

1. Postman → **Import** → select `backend/docs/openapi.yaml` (or drag-and-drop it).
2. Postman will generate a collection with one request per endpoint.
3. Set the collection/environment variable `baseUrl` (or edit the server URL) to your running instance, default `http://localhost:8080/api/v1`.
4. Get a Bearer token: `POST /auth/login` (emails a code), then `POST /auth/login/verify` with that code; set the collection's Authorization to Bearer using `data.AccessTokenStr` from the verify response. All routes except `/auth/*` require it.

## Notes / known quirks (matched to actual server behavior)

- `GET` endpoints return the raw DTO JSON (no envelope); errors are wrapped in a `{info, errors, request_id}` envelope.
- Mutating endpoints (`POST` create, `PATCH`, `DELETE`) **and every `/auth/*` endpoint** return a `{info: {code, key, message}, data, request_id}` success envelope instead of a bare status with no body. `key`/`message` come from `response.SuccessRegistry`, keyed by the domain's `SuccessKey` (e.g. `FARM.CREATE_SUCCESS`). Create endpoints return `201` with the new record's id (or ids) in `data`, so a client never has to list and search for what it just made; update/delete return `200`.
- Auth is a two-step email-OTP flow: `POST /auth/register` and `POST /auth/login` only email a code; the matching `/verify` endpoint returns the token pair (under `data`) and sets the `rid` cookie. `POST /auth/{purpose}/resend-code` (`purpose` = `register` or `reset_password`) re-sends. Password reset is `reset-password/request` then `reset-password/confirm`.
- `/auth/refresh` and `/auth/logout` read the refresh token from an httpOnly `rid` cookie, not the body. `/auth/refresh` returns a rotated pair in `data` but does not itself re-set the `rid` cookie.
- `TokenPair` (verify/refresh response) uses PascalCase field names (`AccessTokenStr`, `RefreshToken.ID`, ...) since the Go struct has no `json` tags. This is inconsistent with the rest of the API but reflects current behavior. It is nested under the success envelope's `data` field.
- `LocationResponse` never includes `farm_id` (excluded via `json:"-"` in the source).
- `TransactionResponse` does not expose the domain's `Currency` field.

If any of these change in the backend, update `openapi.yaml` to match.
