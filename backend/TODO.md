# TODO

_Last gap analysis: 2026-08-28 (round 4). Implementation pass since: see "Fixed since round 4"._

## Security

Nothing open. The three items from round 4 are under "Fixed since round 4".

## Correctness

Nothing open.

## Schema / migrations

Nothing open. Every foreign key declares `ON DELETE CASCADE` (checked across migrations `00001` to
`00007`), so deleting a user, farm, product or post takes its children with it.

## Cleanup

Nothing open.

## Data hygiene

Nothing open.

## Fixed since round 4

- [x] **Login rejects `email` and `username` together** (400) instead of
      letting the second lookup overwrite the first.
- [x] **A password reset verifies the email** if the account never had, since
      the code was sent to that address. Without it the new password still
      couldn't sign in.
- [x] **Auth handlers use `HandleServiceError`** like every other handler. The
      old `WriteError(..., ErrorKey(err.Error()))` also put the raw error text
      in the response's `key`, which for an unexpected error was an internal
      message.
- [x] **Engagement endpoints answer with the standard envelope**
      (`POST.ENGAGEMENT_CREATE_SUCCESS` / `POST.ENGAGEMENT_DELETE_SUCCESS`).
- [x] **`domain/note` is documented as a write-only audit trail.**
- [x] **Removed `ListUsersByAsset`** (interface, repository, fakes and test)
      and the unused `LoginCommand` / `ChangePasswordCommand`.
- [x] **Expired refresh tokens are swept hourly**
      (`Service.PruneExpiredSessions`, started from `cmd/api/main.go`).
- [x] **Login no longer reveals which accounts exist.** An unknown
      email/username and a wrong password both answer
      `AUTH.INVALID_CREDENTIALS` (401), and the unknown-account path runs a
      throwaway bcrypt compare so the two take about as long.
      `AUTH.EMAIL_NOT_VERIFIED` is only returned after a correct password.
- [x] **Password-reset request answers identically for known and unknown
      emails.** A rate-limited or failed send for a known account is logged,
      not returned.
- [x] **Confirming a password reset revokes every session.** The new hash
      and `InvalidateSessions` run in one transaction; the same goes for
      `PATCH /user/{id}/password`, which now writes the password first.
- [x] **Deleting an account signs it out** (`InvalidateSessions` inside the
      delete transaction).
- [x] **Registration works for a new email.** `Register` treated
      `FindUserByEmail`'s `ErrNotFound` (what the Postgres repo returns for
      "nobody has this email") as a failure, so no new account could be
      created. Only the unit tests' fake, which returned an empty user,
      hid it.
- [x] **Soft-deleted users free their email and username.** The whole-table
      unique constraints are now partial unique indexes on rows with
      `deleted_at IS NULL` (migration `00012`).
- [x] **CORS.** `middleware.CORS(cfg.Cors.AllowedOrigins)` wraps the whole
      mux in `cmd/api/main.go` (so it also covers the method-pattern auth
      routes' preflight, which `ServeMux` would 405). Validates `Origin`
      against the `cors.allowed_origins` config list (`"*"` = reflect any,
      dev only), answers preflight `OPTIONS` with
      methods/headers/max-age + `204`, sends
      `Access-Control-Allow-Credentials: true` (auth uses the `rid` cookie),
      and exposes `X-Request-Id`. Requests with no `Origin` (mobile/native,
      server-to-server) pass through untouched. The pre-existing `OPTIONS →
      204` in `routes.go` is now only hit by non-preflight `OPTIONS` probes.
- [x] **`request_id` is now populated.** `middleware.RequestID` (outermost
      in both stacks) generates a UUID per request (inbound `X-Request-Id`
      ignored — backend is the source of truth), stores it under
      `shared.KeyRequestID`, echoes it on the response header, and the logger
      emits it. `WriteError` already used it; `WriteSuccess` now does too.
      (`shared.KeyRequestID` was read everywhere but set nowhere.)
- [x] **Panic-recovery middleware.** `middleware.Recover`
      (`internal/api/middleware/recover.go`) sits outermost in both
      `CreateStack` stacks in `cmd/api/main.go`; it logs the recovered value
      + `debug.Stack()` and writes the standard `COMMON.INTERNAL_SERVER_ERROR`
      envelope, re-panicking `http.ErrAbortHandler`.
- [x] **`GET /users` IDOR + non-functional asset scoping.** `GetUsers`
      dropped the caller-supplied `asset_type` / `asset_id` entirely; it is
      now admin-only, enforced both at the route (`PermGetManyUsers` removed
      from `userPermissions`) and in `app/user.GetUsers`
      (`shared.KeyRole != RoleAdmin` → `shared.ErrForbidden`).
      `userRepo.ListUsers` is implemented
      (`SELECT * FROM users WHERE deleted_at IS NULL ORDER BY created_at`),
      replacing the `panic("implement me")`. The raw
      `errors.New("invalid asset type")` and the identity `switch` in both
      the service and the handler are gone.
- [x] **`PATCH /user/{id}/password` revokes sessions and verifies the
      current password in the service.** `app/user.UpdateUserPassword` now
      takes `(oldPassword, newPassword)`, runs
      `bcrypt.CompareHashAndPassword` against the stored hash (→
      `shared.ErrUnauthorized` on mismatch), then
      `authRepo.InvalidateSessions` before writing the new hash. The handler
      passes `req.OldPassword` through instead of doing the bcrypt check
      itself. (Reset path still open — see Security.)
- [x] **`PATCH /farm/{id}/product/{id}` now changes the product.**
      `UpdateProduct` applies the non-nil `cmd.Name` / `cmd.Description` /
      `cmd.Image` to the loaded entity before `productRepo.UpdateProduct`,
      so the row and the audit note agree.
- [x] **`transactions.currency` is wired end to end.** `TransactionRequest`
      has `Currency string \`validate:"required,iso4217"\``,
      `TransactionCommand` carries it, `ToCommand` and `ToDomain` both set
      it, and `TransactionResponse` / `DomainToTransactionResponse` expose
      it. No more `currency = ''`.

## Resolved in round 4

- [x] `POST /auth/refresh` now calls `setAuthCookie(w, tokenPair)`, so a
      cookie-only client keeps a valid `rid` after rotation.
- [x] Cookie construction extracted to `internal/api/auth/cookie.go`
      (`setAuthCookie` / `removeAuthCookie`), used by `VerifyRegister`,
      `VerifyLogin`, `Refresh`, `Logout` — no more hand-rolled triplication.
- [x] `PATCH /user/{id}/password` now runs `ValidateStruct` +
      `utils.ValidPassword(NewPassword)` in the handler before delegating,
      so weak passwords are rejected.

## Resolved earlier

- [x] Rate limiting behind the documented Traefik proxy: added
      `server.trusted_proxies`; `Middleware.clientIP` walks
      `X-Forwarded-For` right-to-left for trusted peers; rate limiter +
      access log + refresh-token `ip_address` now key on the real client IP.
- [x] `Register` transactionality reworked to stage the password-hashed
      user in Redis and only `CreateUser` from `VerifyRegistration`; code
      consumed only after `CreateUser` + `VerifyUserEmail` succeed; staged
      key namespaced; unique-violation mapped to
      `ErrEmailExists` / `ErrUsernameExists`.
- [x] Removed the broken reset-then-increment rate limiter from the auth
      service.
- [x] `POST /auth/{purpose}/resend-code` rejects `purpose=login`; email
      moved to a validated JSON body.
- [x] `rid` cookie set by `VerifyRegister` / `VerifyLogin` / `Logout`.
- [x] `cmd/email` worker consumes `QueueNameEmail` and sends verification
      emails.
- [x] UNIQUE constraints on `users.email` / `users.username` added in
      `00001`; secondary indexes on FK/lookup columns across `00001`–`00007`.
- [x] `docs/openapi.yaml`, `docs/README.md` and root `README.md` rewritten
      for the two-step email-OTP auth flow and current infra/test setup.
