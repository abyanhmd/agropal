package farm

import (
	"agropal/config"
	aFarm "agropal/internal/app/farm"
	dFarm "agropal/internal/domain/farm"
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
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier; the
// fake repositories below never touch the Querier they're given.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by farm.Service (via verifyUser); everything else panics if
// called.
type fakeUserRepository struct {
	findUserByIDFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
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

func (f *fakeUserRepository) ListUsers(context.Context, database.Querier) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsers not stubbed")
}

func (f *fakeUserRepository) UpdateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: UpdateUser not stubbed")
}

func (f *fakeUserRepository) UpdateUserPassword(context.Context, database.Querier, uuid.UUID, string) error {
	panic("fakeUserRepository: UpdateUserPassword not stubbed")
}

func (f *fakeUserRepository) VerifyUserEmail(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: VerifyUserEmail not stubbed")
}

func (f *fakeUserRepository) SoftDeleteUser(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: SoftDeleteUser not stubbed")
}

func (f *fakeUserRepository) DeleteUser(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeUserRepository: DeleteUser not stubbed")
}

// fakeFarmRepository implements dFarm.Repository. Only stubbed methods are
// callable; everything else panics.
type fakeFarmRepository struct {
	createFarmFunc       func(ctx context.Context, q database.Querier, farm dFarm.Farm) error
	findFarmByIDFunc     func(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error)
	updateFarmFunc       func(ctx context.Context, q database.Querier, farm dFarm.Farm) error
	deleteFarmFunc       func(ctx context.Context, q database.Querier, farmID uuid.UUID) error
	listFarmByUserIDFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error)
}

func (f *fakeFarmRepository) CreateFarm(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	if f.createFarmFunc == nil {
		panic("fakeFarmRepository: createFarmFunc not stubbed")
	}
	return f.createFarmFunc(ctx, q, farm)
}

func (f *fakeFarmRepository) FindFarmByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error) {
	if f.findFarmByIDFunc == nil {
		panic("fakeFarmRepository: findFarmByIDFunc not stubbed")
	}
	return f.findFarmByIDFunc(ctx, q, farmID)
}

func (f *fakeFarmRepository) UpdateFarm(ctx context.Context, q database.Querier, farm dFarm.Farm) error {
	if f.updateFarmFunc == nil {
		panic("fakeFarmRepository: updateFarmFunc not stubbed")
	}
	return f.updateFarmFunc(ctx, q, farm)
}

func (f *fakeFarmRepository) DeleteFarm(ctx context.Context, q database.Querier, farmID uuid.UUID) error {
	if f.deleteFarmFunc == nil {
		panic("fakeFarmRepository: deleteFarmFunc not stubbed")
	}
	return f.deleteFarmFunc(ctx, q, farmID)
}

func (f *fakeFarmRepository) ListFarmByUserID(ctx context.Context, q database.Querier, userID uuid.UUID) ([]dFarm.Farm, error) {
	if f.listFarmByUserIDFunc == nil {
		panic("fakeFarmRepository: listFarmByUserIDFunc not stubbed")
	}
	return f.listFarmByUserIDFunc(ctx, q, userID)
}

/* ==================== Helpers ====================== */

func newTestHandler(userRepo dUser.Repository, farmRepo dFarm.Repository) *Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aFarm.NewService(fakeTransactor{}, userRepo, farmRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())

	h := NewHandler(ioHandler, svc)
	return &h
}

// authedRequest returns req with a context carrying userID the way
// middleware.RequireAuth injects it after JWT validation, via
// shared.SaveContextFromRequest.
func authedRequest(req *http.Request, userID uuid.UUID, role string) *http.Request {
	claims := shared.ClaimInfo{UserID: userID, Role: role}
	ctx := shared.SaveContextFromRequest(req, &claims)
	return req.WithContext(ctx)
}

func jsonRequest(method, path, body string) *http.Request {
	return httptest.NewRequest(method, path, bytes.NewBufferString(body))
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) utils.StandardEnvelope {
	t.Helper()
	var envelope utils.StandardEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
	return envelope
}

/* ==================== Tests: Create ====================== */

func TestHandler_Create(t *testing.T) {
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			createFarmFunc: func(context.Context, database.Querier, dFarm.Farm) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(jsonRequest(http.MethodPost, "/farm", `{"name":"Green Acres"}`), userID, "user")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dFarm.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dFarm.SuccessCreated.String())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{})

		req := authedRequest(jsonRequest(http.MethodPost, "/farm", "not-json"), userID, "user")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Missing Required Name Fails Validation Before Hitting The Service", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				t.Fatal("service should not be called when validation fails")
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeFarmRepository{})

		req := authedRequest(jsonRequest(http.MethodPost, "/farm", `{"name":""}`), userID, "user")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrValidation.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrValidation.Error())
		}
	})

	t.Run("Service Error Maps Through HandleServiceError", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeFarmRepository{})

		req := authedRequest(jsonRequest(http.MethodPost, "/farm", `{"name":"Green Acres"}`), userID, "user")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dUser.ErrInvalidUserID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dUser.ErrInvalidUserID.Error())
		}
	})
}

/* ==================== Tests: GetOne ====================== */

func TestHandler_GetOne(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID, Name: "Green Acres"}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String(), nil), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got response.FarmResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if got.FarmID != farmID || got.Name != "Green Acres" {
			t.Errorf("response = %+v, want FarmID=%v Name=%q", got, farmID, "Green Acres")
		}
	})

	t.Run("Invalid Path ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{})

		req := authedRequest(httptest.NewRequest(http.MethodGet, "/farm/not-a-uuid", nil), userID, "user")
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dFarm.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dFarm.ErrInvalidID.Error())
		}
	})

	t.Run("Belonging To Another User Maps To 403", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: uuid.New(), Name: "Not Mine"}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String(), nil), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: GetMany ====================== */

func TestHandler_GetMany(t *testing.T) {
	userID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			listFarmByUserIDFunc: func(context.Context, database.Querier, uuid.UUID) ([]dFarm.Farm, error) {
				return []dFarm.Farm{{ID: uuid.New(), UserID: userID, Name: "Farm A"}}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(httptest.NewRequest(http.MethodGet, "/farm", nil), userID, "user")
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []response.FarmResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if len(got) != 1 || got[0].Name != "Farm A" {
			t.Errorf("response = %+v, want one farm named Farm A", got)
		}
	})

	t.Run("Service Error Maps Through HandleServiceError", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(userRepo, &fakeFarmRepository{})

		req := authedRequest(httptest.NewRequest(http.MethodGet, "/farm", nil), userID, "user")
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

/* ==================== Tests: Update ====================== */

func TestHandler_Update(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID, Name: "Old Name"}, nil
			},
			updateFarmFunc: func(context.Context, database.Querier, dFarm.Farm) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(jsonRequest(http.MethodPatch, "/farm/"+farmID.String(), `{"name":"New Name"}`), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dFarm.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dFarm.SuccessUpdated.String())
		}
	})

	t.Run("Invalid Path ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{})

		req := authedRequest(jsonRequest(http.MethodPatch, "/farm/not-a-uuid", `{"name":"New Name"}`), userID, "user")
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("Missing Required Name Fails Validation", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{})

		req := authedRequest(jsonRequest(http.MethodPatch, "/farm/"+farmID.String(), `{"name":""}`), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrValidation.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrValidation.Error())
		}
	})

	t.Run("Belonging To Another User Without Admin Role Maps To 403", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: uuid.New(), Name: "Not Mine"}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(jsonRequest(http.MethodPatch, "/farm/"+farmID.String(), `{"name":"New Name"}`), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: Delete ====================== */

func TestHandler_Delete(t *testing.T) {
	userID := uuid.New()
	farmID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{ID: farmID, UserID: userID, Name: "Green Acres"}, nil
			},
			deleteFarmFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String(), nil), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dFarm.SuccessDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dFarm.SuccessDeleted.String())
		}
	})

	t.Run("Invalid Path ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{})

		req := authedRequest(httptest.NewRequest(http.MethodDelete, "/farm/not-a-uuid", nil), userID, "user")
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("Not Found Maps Through HandleServiceError", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return dFarm.Farm{}, dFarm.ErrNotFound
			},
		}
		h := newTestHandler(userRepo, farmRepo)

		req := authedRequest(httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String(), nil), userID, "user")
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dFarm.ErrNotFound.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dFarm.ErrNotFound.Error())
		}
	})
}
