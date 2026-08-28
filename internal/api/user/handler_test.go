package user

import (
	"agropal/config"
	aUser "agropal/internal/app/user"
	dAuth "agropal/internal/domain/auth"
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/response"
	"agropal/internal/utils"
	"agropal/pkg"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/* ==================== Fakes ====================== */

type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

type fakeUserRepository struct {
	findUserByIDFunc       func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
	updateUserFunc         func(ctx context.Context, q database.Querier, user dUser.User) error
	updateUserPasswordFunc func(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error
	softDeleteUserFunc     func(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error
	deleteUserFunc         func(ctx context.Context, q database.Querier, userID uuid.UUID) error
	listUsersFunc          func(ctx context.Context, q database.Querier) ([]dUser.User, error)
}

func (f *fakeUserRepository) CreateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: CreateUser not stubbed")
}

func (f *fakeUserRepository) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	if f.findUserByIDFunc == nil {
		panic("fakeUserRepository: findUserByIDFunc not stubbed")
	}
	return f.findUserByIDFunc(ctx, q, userID)
}

func (f *fakeUserRepository) FindUserByEmail(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByEmail not stubbed")
}

func (f *fakeUserRepository) FindUserByUsername(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByUsername not stubbed")
}

func (f *fakeUserRepository) ListUsers(ctx context.Context, q database.Querier) ([]dUser.User, error) {
	if f.listUsersFunc == nil {
		panic("fakeUserRepository: listUsersFunc not stubbed")
	}
	return f.listUsersFunc(ctx, q)
}

func (f *fakeUserRepository) UpdateUser(ctx context.Context, q database.Querier, user dUser.User) error {
	if f.updateUserFunc == nil {
		panic("fakeUserRepository: updateUserFunc not stubbed")
	}
	return f.updateUserFunc(ctx, q, user)
}

func (f *fakeUserRepository) UpdateUserPassword(ctx context.Context, q database.Querier, userID uuid.UUID, password string) error {
	if f.updateUserPasswordFunc == nil {
		panic("fakeUserRepository: updateUserPasswordFunc not stubbed")
	}
	return f.updateUserPasswordFunc(ctx, q, userID, password)
}

func (f *fakeUserRepository) VerifyUserEmail(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: VerifyUserEmail not stubbed")
}

func (f *fakeUserRepository) SoftDeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID, deletedAt time.Time) error {
	if f.softDeleteUserFunc == nil {
		panic("fakeUserRepository: softDeleteUserFunc not stubbed")
	}
	return f.softDeleteUserFunc(ctx, q, userID, deletedAt)
}

func (f *fakeUserRepository) DeleteUser(ctx context.Context, q database.Querier, userID uuid.UUID) error {
	if f.deleteUserFunc == nil {
		panic("fakeUserRepository: deleteUserFunc not stubbed")
	}
	return f.deleteUserFunc(ctx, q, userID)
}

func (f *fakeUserRepository) ListUsersByAsset(context.Context, database.Querier, shared.AssetType, uuid.UUID) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsersByAsset not stubbed")
}

// fakeAuthRepository implements dAuth.Repository; user.Service only calls
// InvalidateSessions (on a password change), which defaults to a no-op here.
type fakeAuthRepository struct{}

func (f *fakeAuthRepository) CreateRefreshToken(context.Context, database.Querier, dAuth.RefreshToken) error {
	panic("fakeAuthRepository: CreateRefreshToken not stubbed")
}

func (f *fakeAuthRepository) GetRefreshToken(context.Context, database.Querier, string) (dAuth.RefreshToken, error) {
	panic("fakeAuthRepository: GetRefreshToken not stubbed")
}

func (f *fakeAuthRepository) DeleteRefreshToken(context.Context, database.Querier, string) error {
	panic("fakeAuthRepository: DeleteRefreshToken not stubbed")
}

func (f *fakeAuthRepository) InvalidateSessions(context.Context, database.Querier, uuid.UUID) error {
	return nil
}

// fakeNoteRepository implements dNote.Repository.
type fakeNoteRepository struct {
	createNoteFunc func(ctx context.Context, q database.Querier, note dNote.Note) error
}

func (f *fakeNoteRepository) CreateNote(ctx context.Context, q database.Querier, note dNote.Note) error {
	if f.createNoteFunc == nil {
		panic("fakeNoteRepository: createNoteFunc not stubbed")
	}
	return f.createNoteFunc(ctx, q, note)
}

func (f *fakeNoteRepository) CreateManyNotes(context.Context, database.Querier, []dNote.Note) error {
	panic("fakeNoteRepository: CreateManyNotes not stubbed")
}

func (f *fakeNoteRepository) ListNotes(context.Context, database.Querier, dNote.NoteType, []uuid.UUID, map[string]string) ([]dNote.Note, error) {
	panic("fakeNoteRepository: ListNotes not stubbed")
}

func (f *fakeNoteRepository) DeleteNotes(context.Context, database.Querier, []uuid.UUID) error {
	panic("fakeNoteRepository: DeleteNotes not stubbed")
}

/* ==================== Helpers ====================== */

func newTestHandler(userRepo dUser.Repository, noteRepo dNote.Repository) *Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aUser.NewService(fakeTransactor{}, userRepo, &fakeAuthRepository{}, noteRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())
	h := NewHandler(ioHandler, svc)
	return &h
}

// authedRequest attaches the metadata the JWT middleware would normally set
// on the request context: the caller's user ID and role.
func authedRequest(req *http.Request, callerID uuid.UUID, role config.Role) *http.Request {
	ctx := req.Context()
	ctx = shared.WithMetadata(ctx, shared.KeyUserID, callerID.String())
	ctx = shared.WithMetadata(ctx, shared.KeyRole, string(role))
	return req.WithContext(ctx)
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) utils.StandardEnvelope {
	t.Helper()
	var envelope utils.StandardEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
	return envelope
}

func newRequestWithPathValue(method, path, pathKey, pathValue, body string) *http.Request {
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	req.SetPathValue(pathKey, pathValue)
	return req
}

/* ==================== Tests: GetByID ====================== */

func TestHandler_GetByID(t *testing.T) {
	t.Run("Self Success", func(t *testing.T) {
		callerID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID, Account: dUser.AccountInfo{Name: "Jane", Email: "jane@example.com", Username: "jane"}}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodGet, "/user/me", "user_id", "me", "")
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got response.UserResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if got.Email != "jane@example.com" {
			t.Errorf("Email = %q, want %q", got.Email, "jane@example.com")
		}
	})

	t.Run("Invalid Path ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodGet, "/user/not-a-uuid", "user_id", "not-a-uuid", "")
		req = authedRequest(req, uuid.New(), config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dUser.ErrInvalidUserID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dUser.ErrInvalidUserID.Error())
		}
	})

	t.Run("Accessing Another User Without Admin Role Maps To 403", func(t *testing.T) {
		callerID := uuid.New()
		targetID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: targetID}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodGet, "/user/"+targetID.String(), "user_id", targetID.String(), "")
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})

	t.Run("Admin Accessing Another User Succeeds", func(t *testing.T) {
		callerID := uuid.New()
		targetID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: targetID, Account: dUser.AccountInfo{Email: "target@example.com"}}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodGet, "/user/"+targetID.String(), "user_id", targetID.String(), "")
		req = authedRequest(req, callerID, config.RoleAdmin)
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("Missing Auth Context Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodGet, "/user/me", "user_id", "me", "")
		rec := httptest.NewRecorder()

		h.GetByID(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

/* ==================== Tests: Update ====================== */

func TestHandler_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		callerID := uuid.New()
		var updated dUser.User
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID}, nil
			},
			updateUserFunc: func(_ context.Context, _ database.Querier, u dUser.User) error {
				updated = u
				return nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me", "user_id", "me", `{"name":"New Name"}`)
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if updated.Account.Name != "New Name" {
			t.Errorf("updated.Account.Name = %q, want %q", updated.Account.Name, "New Name")
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dUser.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dUser.SuccessUpdated.String())
		}
	})

	t.Run("Invalid JSON Body Maps To 400", func(t *testing.T) {
		callerID := uuid.New()
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me", "user_id", "me", "not-json")
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Invalid Path ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/not-a-uuid", "user_id", "not-a-uuid", `{"name":"x"}`)
		req = authedRequest(req, uuid.New(), config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

/* ==================== Tests: ChangePassword ====================== */

func TestHandler_ChangePassword(t *testing.T) {
	callerID := uuid.New()
	oldHashed, err := bcrypt.GenerateFromPassword([]byte("OldStr0ng!Pass"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword() error = %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		var newPassword string
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID, Security: dUser.SecurityInfo{Password: string(oldHashed)}}, nil
			},
			updateUserPasswordFunc: func(_ context.Context, _ database.Querier, _ uuid.UUID, password string) error {
				newPassword = password
				return nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me/password", "user_id", "me",
			`{"old_password":"OldStr0ng!Pass","new_password":"NewStr0ng!Pass1"}`)
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if newPassword == "" || newPassword == "NewStr0ng!Pass1" {
			t.Errorf("stored password = %q, want a bcrypt hash of the new password", newPassword)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(newPassword), []byte("NewStr0ng!Pass1")); err != nil {
			t.Errorf("stored password does not verify against the new password: %v", err)
		}
	})

	t.Run("Wrong Old Password Maps To 401", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID, Security: dUser.SecurityInfo{Password: string(oldHashed)}}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me/password", "user_id", "me",
			`{"old_password":"wrong-password","new_password":"NewStr0ng!Pass1"}`)
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})

	t.Run("Missing Required Field Fails Struct Validation", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me/password", "user_id", "me",
			`{"old_password":"OldStr0ng!Pass"}`)
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("Weak New Password Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodPatch, "/user/me/password", "user_id", "me",
			`{"old_password":"OldStr0ng!Pass","new_password":"short"}`)
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.ChangePassword(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

/* ==================== Tests: Delete ====================== */

func TestHandler_Delete(t *testing.T) {
	t.Run("Soft Delete Success (default)", func(t *testing.T) {
		callerID := uuid.New()
		var softDeleted, hardDeleted bool
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID}, nil
			},
			softDeleteUserFunc: func(context.Context, database.Querier, uuid.UUID, time.Time) error {
				softDeleted = true
				return nil
			},
			deleteUserFunc: func(context.Context, database.Querier, uuid.UUID) error {
				hardDeleted = true
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(context.Context, database.Querier, dNote.Note) error { return nil },
		}
		h := newTestHandler(userRepo, noteRepo)

		req := newRequestWithPathValue(http.MethodDelete, "/user/me", "user_id", "me", "")
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !softDeleted || hardDeleted {
			t.Errorf("softDeleted=%v hardDeleted=%v, want soft delete only", softDeleted, hardDeleted)
		}
	})

	t.Run("Permanent Delete Success", func(t *testing.T) {
		callerID := uuid.New()
		var hardDeleted bool
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: callerID}, nil
			},
			deleteUserFunc: func(context.Context, database.Querier, uuid.UUID) error {
				hardDeleted = true
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(context.Context, database.Querier, dNote.Note) error { return nil },
		}
		h := newTestHandler(userRepo, noteRepo)

		req := newRequestWithPathValue(http.MethodDelete, "/user/me?permanent=true", "user_id", "me", "")
		req.URL.RawQuery = "permanent=true"
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if !hardDeleted {
			t.Error("expected DeleteUser (hard delete) to be called")
		}
	})

	t.Run("Invalid Permanent Query Value Maps To 400", func(t *testing.T) {
		callerID := uuid.New()
		h := newTestHandler(&fakeUserRepository{}, &fakeNoteRepository{})

		req := newRequestWithPathValue(http.MethodDelete, "/user/me", "user_id", "me", "")
		req.URL.RawQuery = "permanent=maybe"
		req = authedRequest(req, callerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

/* ==================== Tests: GetUsers ====================== */

func TestHandler_GetUsers(t *testing.T) {
	t.Run("Admin Success", func(t *testing.T) {
		adminID := uuid.New()
		userRepo := &fakeUserRepository{
			listUsersFunc: func(context.Context, database.Querier) ([]dUser.User, error) {
				return []dUser.User{{ID: uuid.New(), Account: dUser.AccountInfo{Email: "a@example.com"}}}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req = authedRequest(req, adminID, config.RoleAdmin)
		rec := httptest.NewRecorder()

		h.GetUsers(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []response.UserResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode body: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("len(got) = %d, want 1", len(got))
		}
	})

	t.Run("Non-Admin Maps To 403", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			listUsersFunc: func(context.Context, database.Querier) ([]dUser.User, error) {
				t.Fatal("ListUsers should not be called for a non-admin caller")
				return nil, nil
			},
		}
		h := newTestHandler(userRepo, &fakeNoteRepository{})

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req = authedRequest(req, uuid.New(), config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetUsers(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}
