# Unused assets

_Scanned 2026-10-24. Nothing here has been removed yet._

How this list was made:

- **Functions:** `go run golang.org/x/tools/cmd/deadcode@v0.38.0 ./cmd/...` (unreachable from `cmd/api`, `cmd/email`, `cmd/migration`), re-run with `-test` to see which are only kept alive by tests.
- **Repository methods:** methods on `internal/domain/*/repository.go` interfaces with no `.Method(` call outside tests. `deadcode` can't see these, since the interface keeps every implementation "reachable".
- **Keys:** every `shared.ErrorKey` / `shared.SuccessKey` declared under `internal/domain` with no reference outside its declaration and the `internal/response/registry_*.go` maps. "Tests" counts references in `_test.go` files: those tests stub or assert the key, so they need updating if the key is removed.

## Functions

| Function | Where | Notes |
|---|---|---|
| `Service.Authorize` | `internal/app/auth/service.go:515` | Only called from `service_test.go`. Role checks are done by `middleware.RequirePermission` instead. |
| `NewRepository` (+ `RepositoryImpl`) | `internal/infra/persistence/persistence.go:13` | Never wired to any SQL implementation; `main_test.go` says so in a comment. |
| `NewResponseRegistry` | `internal/response/registry.go:24` | Part of an unused YAML-backed registry; the real registries are the Go maps in `registry_error.go` / `registry_success.go`. |
| `LoadRegistries` | `internal/response/registry.go:34` | Same. |
| `loadFile` | `internal/response/registry.go:51` | Same. |
| `ResponseRegistry.GetErrorRegistry` | `internal/response/registry.go:65` | Same. |
| `ResponseRegistry.GetSuccessRegistry` | `internal/response/registry.go:69` | Same. |
| `ResponseRegistry.lookup` | `internal/response/registry.go:73` | Same. |

Removing the registry functions leaves `ResponseRegistry` / `Registry` referenced only by the `IOHandler.Registry` field (`internal/utils/json.go:16`). `ResponseDetail` is still used by the live maps and must stay.

## Repository methods

Each is declared on the interface, implemented in `internal/infra/persistence`, and implemented again by the fakes in tests.

| Method | Interface | Implementation |
|---|---|---|
| `ListNotes` | `internal/domain/note/repository.go:13` | `persistence/note_repo.go:80` |
| `DeleteNotes` | `internal/domain/note/repository.go:14` | `persistence/note_repo.go:125` |
| `UpdatePost` | `internal/domain/post/repository.go:12` | `persistence/post_repo.go:44` |
| `DeletePost` | `internal/domain/post/repository.go:13` | `persistence/post_repo.go:61` |
| `FindPostBySlug` | `internal/domain/post/repository.go:16` | `persistence/post_repo.go:95` |
| `ListTopLevelComments` | `internal/domain/post/repository.go:20` | `persistence/post_repo.go:181` |
| `ListUsersByAsset` | `internal/domain/user/repository.go:25` | `persistence/user_repo.go:230` |

`ListNotes` and `FindPostBySlug` are what a notes read API and a by-slug post lookup would need (see `TODO.md` and the frontend's slug-based discussion URLs), so check before deleting those two.

## Error and success keys

### ⚠ Probably should be used rather than deleted

These not-found keys are never returned: the farm, product, location, transaction and post repositories don't map `sql.ErrNoRows` to them (only `user_repo.go` does, and `note_repo.go` returns its own not-found key), so a missing row likely surfaces as a 500 instead of a 404. Tests stub them, which hides the gap.

| Key | Declared | Tests |
|---|---|---|
| `FARM.NOT_FOUND` | `internal/domain/farm/errors.go:8` | 5 |
| `LOCATION.NOT_FOUND` | `internal/domain/location/errors.go:8` | 2 |
| `POST.NOT_FOUND` | `internal/domain/post/errors.go:8` | 1 |
| `POST.COMMENT_NOT_FOUND` | `internal/domain/post/errors.go:22` | 0 |
| `PRODUCT.NOT_FOUND` | `internal/domain/product/errors.go:8` | 0 |
| `PRODUCT.STOCK_NOT_FOUND` | `internal/domain/product/errors.go:21` | 0 |
| `PRODUCT.PRICE_NOT_FOUND` | `internal/domain/product/errors.go:26` | 0 |
| `TRANSACTION.NOT_FOUND` | `internal/domain/transaction/errors.go:8` | 0 |

### Auth — `internal/domain/auth/errors.go`

| Key | Line |
|---|---|
| `AUTH.INVALID_AVATAR` | 11 |
| `AUTH.WEAK_PASSWORD` | 12 |
| `AUTH.PASSWORD_MISMATCH` | 13 |
| `AUTH.INVALID_INVITE_CODE` | 20 |
| `AUTH.ACCOUNT_DISABLED` | 28 |
| `AUTH.ACCOUNT_LOCKED` | 29 |
| `AUTH.ACCOUNT_NOT_FOUND` | 30 |
| `AUTH.LOGIN_REQUIRED` | 31 |
| `AUTH.TOKEN_EXPIRED` | 34 |
| `AUTH.SESSION_EXPIRED` | 37 |
| `AUTH.SESSION_INVALID` | 38 |
| `AUTH.INSUFFICIENT_SCOPE` | 42 |
| `AUTH.ROLE_REQUIRED` | 43 |
| `AUTH.PASSWORD_RESET_REQUIRED` | 45 |
| `AUTH.PASSWORD_RESET_EXPIRED` | 46 |
| `AUTH.PASSWORD_RESET_INVALID` | 47 |
| `AUTH.OLD_PASSWORD_MISMATCH` | 48 |
| `AUTH.SUSPICIOUS_LOGIN` | 52 |

### Farm — `internal/domain/farm/errors.go`

| Key | Line |
|---|---|
| `FARM.INVALID_NAME` | 11 |
| `FARM.CREATE_FAILED` | 14 |
| `FARM.UPDATE_FAILED` | 15 |
| `FARM.DELETE_FAILED` | 16 |

### Location — `internal/domain/location/errors.go`

| Key | Line |
|---|---|
| `LOCATION.INVALID_LONGITUDE` | 12 |
| `LOCATION.INVALID_LATITUDE` | 13 |
| `LOCATION.CREATE_FAILED` | 15 |
| `LOCATION.UPDATE_FAILED` | 16 |
| `LOCATION.DELETE_FAILED` | 17 |
| `LOCATION.SERVICE_UNAVAILABLE` | 19 |
| `LOCATION.DATA_FETCH_FAILED` | 20 |

Coordinate ranges are checked by `validate` tags on the request DTO, which report `COMMON.VALIDATION_FAILED`, so the two `INVALID_*` keys never fire.

### Note — `internal/domain/note/`

| Key | Declared | Tests |
|---|---|---|
| `NOTE.CREATE_SUCCESS` | `entity.go:38` | 4 |
| `NOTE.UPDATE_SUCCESS` | `entity.go:39` | 4 |
| `NOTE.DELETE_SUCCESS` | `entity.go:40` | 2 |
| `NOTE.MISSING_NOTE_ID` | `errors.go:10` | 0 |
| `NOTE.INVALID_NOTE_ID` | `errors.go:11` | 0 |
| `NOTE.INVALID_TYPE` | `errors.go:12` | 0 |
| `NOTE.INVALID_ASSET_TYPE` | `errors.go:13` | 0 |
| `NOTE.INVALID_AUTHOR_ID` | `errors.go:15` | 0 |
| `NOTE.CREATE_FAILED` | `errors.go:17` | 1 |
| `NOTE.UPDATE_FAILED` | `errors.go:18` | 0 |
| `NOTE.DELETE_FAILED` | `errors.go:19` | 0 |

There is no note API, so none of these can reach a client.

### Post — `internal/domain/post/`

| Key | Declared | Tests |
|---|---|---|
| `POST.UPDATE_SUCCESS` | `entity.go:78` | 0 |
| `POST.DELETE_SUCCESS` | `entity.go:79` | 0 |
| `POST.COMMENT_UPDATE_SUCCESS` | `entity.go:82` | 0 |
| `POST.COMMENT_DELETE_SUCCESS` | `entity.go:83` | 0 |
| `POST.INVALID_TITLE` | `errors.go:11` | 0 |
| `POST.INVALID_CONTENT` | `errors.go:12` | 0 |
| `POST.INVALID_SLUG` | `errors.go:13` | 0 |
| `POST.INVALID_AUTHOR_ID` | `errors.go:14` | 0 |
| `POST.INVALID_ASSET_TYPE` | `errors.go:15` | 0 |
| `POST.CREATE_FAILED` | `errors.go:18` | 1 |
| `POST.UPDATE_FAILED` | `errors.go:19` | 0 |
| `POST.DELETE_FAILED` | `errors.go:20` | 0 |
| `POST.INVALID_PARENT_ID` | `errors.go:24` | 0 |

The update/delete keys match the unused `UpdatePost` / `DeletePost` repository methods: there are no routes for them.

### Product — `internal/domain/product/errors.go`

| Key | Line |
|---|---|
| `PRODUCT.INVALID_NAME` | 11 |
| `PRODUCT.INVALID_DESCRIPTION` | 12 |
| `PRODUCT.INVALID_IMAGE` | 13 |
| `PRODUCT.CREATE_FAILED` | 17 |
| `PRODUCT.UPDATE_FAILED` | 18 |
| `PRODUCT.DELETE_FAILED` | 19 |
| `PRODUCT.INVALID_STOCK_TYPE` | 23 |
| `PRODUCT.INSUFFICIENT_STOCK` | 30 |
| `PRODUCT.INVALID_MOVEMENT_TYPE` | 31 |

### Transaction — `internal/domain/transaction/errors.go`

| Key | Line |
|---|---|
| `TRANSACTION.INVALID_TYPE` | 11 |
| `TRANSACTION.INVALID_AMOUNT` | 12 |
| `TRANSACTION.INVALID_CURRENCY` | 13 |
| `TRANSACTION.INVALID_DESCRIPTION` | 14 |
| `TRANSACTION.CREATE_FAILED` | 17 |
| `TRANSACTION.UPDATE_FAILED` | 18 |
| `TRANSACTION.DELETE_FAILED` | 19 |

### User — `internal/domain/user/errors.go`

| Key | Line | Tests |
|---|---|---|
| `USER.INVALID_NAME` | 13 | 0 |
| `USER.INVALID_EMAIL` | 14 | 2 |
| `USER.INVALID_USERNAME` | 15 | 0 |
| `USER.INVALID_AVATAR` | 16 | 0 |
| `USER.INVALID_ROLE` | 18 | 0 |
| `USER.UPDATE_FAILED` | 20 | 0 |
| `USER.DELETE_FAILED` | 21 | 0 |

## Removing an entry

For a key: delete the declaration, its line in `internal/response/registry_error.go` or `registry_success.go`, and any test that stubs it. For a repository method: delete it from the interface, the `persistence` implementation, its repository test, and every fake that implements the interface (`grep -rn "fake.*Repository" internal`).
