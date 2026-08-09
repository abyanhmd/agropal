# Agropal API: OpenAPI Spec

`openapi.yaml` is an OpenAPI 3.0.3 spec covering all HTTP routes exposed by `cmd/api`.

## Import into Postman

1. Postman → **Import** → select `backend/docs/openapi.yaml` (or drag-and-drop it).
2. Postman will generate a collection with one request per endpoint.
3. Set the collection/environment variable `baseUrl` (or edit the server URL) to your running instance, default `http://localhost:8080/api/v1`.
4. Set a Bearer token: call `POST /auth/login`, then set the collection's Authorization to Bearer using `AccessTokenStr` from the response. All routes except `/auth/*` require it.

## Notes / known quirks (matched to actual server behavior)

- `GET` endpoints (and `POST /auth/login`/`/auth/refresh`) return the raw DTO JSON (no envelope); errors are wrapped in a `{info, errors, request_id}` envelope.
- Mutating endpoints (`POST` create, `PATCH`, `DELETE`) return a `{info: {code, key, message}, data, request_id}` success envelope instead of a bare status with no body. `key`/`message` come from `response.SuccessRegistry`, keyed by the domain's `SuccessKey` (e.g. `FARM.CREATE_SUCCESS`). Create endpoints return `201`, update/delete return `200`. The `/post/{post_id}/engagement` and `/comment/{comment_id}/engagement` endpoints are the current exception: they still return `201` with no body since no success key is registered for them yet.
- `/auth/refresh` and `/auth/logout` read the refresh token from an httpOnly `rid` cookie, not the body.
- `TokenPair` (login/refresh response) uses PascalCase field names (`AccessTokenStr`, `RefreshToken.ID`, ...) since the Go struct has no `json` tags. This is inconsistent with the rest of the API but reflects current behavior. It's now nested under the success envelope's `data` field rather than being the top-level response body.
- `WeatherResponse` never includes `farm_id` (excluded via `json:"-"` in the source).
- `TransactionResponse` does not expose the domain's `Currency` field.

If any of these change in the backend, update `openapi.yaml` to match.
