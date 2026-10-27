package location

import (
	"agropal/config"
	aLocation "agropal/internal/app/location"
	dFarm "agropal/internal/domain/farm"
	dLocation "agropal/internal/domain/location"
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

type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by the location service; everything else panics if called.
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

// fakeFarmRepository implements dFarm.Repository. Only FindFarmByID is
// exercised by the location service.
type fakeFarmRepository struct {
	findFarmByIDFunc func(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error)
}

func (f *fakeFarmRepository) CreateFarm(context.Context, database.Querier, dFarm.Farm) error {
	panic("fakeFarmRepository: CreateFarm not stubbed")
}
func (f *fakeFarmRepository) FindFarmByID(ctx context.Context, q database.Querier, farmID uuid.UUID) (dFarm.Farm, error) {
	if f.findFarmByIDFunc == nil {
		panic("fakeFarmRepository: findFarmByIDFunc not stubbed")
	}
	return f.findFarmByIDFunc(ctx, q, farmID)
}
func (f *fakeFarmRepository) UpdateFarm(context.Context, database.Querier, dFarm.Farm) error {
	panic("fakeFarmRepository: UpdateFarm not stubbed")
}
func (f *fakeFarmRepository) DeleteFarm(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeFarmRepository: DeleteFarm not stubbed")
}
func (f *fakeFarmRepository) ListFarmByUserID(context.Context, database.Querier, uuid.UUID) ([]dFarm.Farm, error) {
	panic("fakeFarmRepository: ListFarmByUserID not stubbed")
}

// fakeLocationRepository implements dLocation.Repository.
type fakeLocationRepository struct {
	createLocationFunc      func(ctx context.Context, q database.Querier, location dLocation.Location) error
	getLocationByIDFunc     func(ctx context.Context, q database.Querier, id uuid.UUID) (dLocation.Location, error)
	listLocationsByFarmFunc func(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dLocation.Location, error)
	updateLocationFunc      func(ctx context.Context, q database.Querier, location dLocation.Location) error
	deleteLocationFunc      func(ctx context.Context, q database.Querier, id uuid.UUID) error
}

func (f *fakeLocationRepository) CreateLocation(ctx context.Context, q database.Querier, location dLocation.Location) error {
	if f.createLocationFunc == nil {
		panic("fakeLocationRepository: createLocationFunc not stubbed")
	}
	return f.createLocationFunc(ctx, q, location)
}
func (f *fakeLocationRepository) GetLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (dLocation.Location, error) {
	if f.getLocationByIDFunc == nil {
		panic("fakeLocationRepository: getLocationByIDFunc not stubbed")
	}
	return f.getLocationByIDFunc(ctx, q, id)
}
func (f *fakeLocationRepository) ListLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dLocation.Location, error) {
	if f.listLocationsByFarmFunc == nil {
		panic("fakeLocationRepository: listLocationsByFarmFunc not stubbed")
	}
	return f.listLocationsByFarmFunc(ctx, q, farmID)
}
func (f *fakeLocationRepository) UpdateLocation(ctx context.Context, q database.Querier, location dLocation.Location) error {
	if f.updateLocationFunc == nil {
		panic("fakeLocationRepository: updateLocationFunc not stubbed")
	}
	return f.updateLocationFunc(ctx, q, location)
}
func (f *fakeLocationRepository) DeleteLocation(ctx context.Context, q database.Querier, id uuid.UUID) error {
	if f.deleteLocationFunc == nil {
		panic("fakeLocationRepository: deleteLocationFunc not stubbed")
	}
	return f.deleteLocationFunc(ctx, q, id)
}

/* ==================== Helpers ====================== */

func newTestHandler(userRepo dUser.Repository, farmRepo dFarm.Repository, locationRepo dLocation.Repository) Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aLocation.NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())
	return NewHandler(ioHandler, svc)
}

// authedRequest returns req with an authenticated caller's userID (and
// optionally role) attached to its context, the way middleware.Middleware
// does after JWT auth succeeds.
func authedRequest(req *http.Request, userID uuid.UUID, role config.Role) *http.Request {
	ctx := shared.WithMetadata(req.Context(), shared.KeyUserID, userID.String())
	ctx = shared.WithMetadata(ctx, shared.KeyRole, string(role))
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
	ownerID := uuid.New()
	farmID := uuid.New()

	t.Run("Unauthenticated Request Maps To 401", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/location", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusUnauthorized, rec.Body.String())
		}
	})

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{createLocationFunc: func(context.Context, database.Querier, dLocation.Location) error {
			return nil
		}}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/location", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.SuccessCreated.String())
		}
	})

	t.Run("Invalid Farm ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		req := jsonRequest(http.MethodPost, "/farm/not-a-uuid/location", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.ErrInvalidFarmID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.ErrInvalidFarmID.Error())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/location", "not-json")
		req.SetPathValue("farm_id", farmID.String())
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

	t.Run("Out Of Range Coordinates Fail Struct Validation", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		// (0, 0) is a valid coordinate, so the "longitude"/"latitude" tags
		// only reject values outside their real-world range.
		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/location", `{"longitude":200,"latitude":200}`)
		req.SetPathValue("farm_id", farmID.String())
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrValidation.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrValidation.Error())
		}
		if len(envelope.Errors) == 0 {
			t.Error("expected per-field validation errors, got none")
		}
	})

	t.Run("Foreign Farm Maps To Forbidden", func(t *testing.T) {
		otherOwnerID := uuid.New()
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: otherOwnerID}, nil
		}}
		h := newTestHandler(userRepo, farmRepo, &fakeLocationRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/location", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: GetByFarmID ====================== */

func TestHandler_GetByFarmID(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		locationID := uuid.New()
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{listLocationsByFarmFunc: func(context.Context, database.Querier, uuid.UUID) ([]dLocation.Location, error) {
			return []dLocation.Location{{LocationID: locationID, FarmID: farmID, Longitude: 1, Latitude: 2}}, nil
		}}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String()+"/location", nil)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByFarmID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []response.LocationResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if len(got) != 1 || got[0].LocationID != locationID {
			t.Errorf("got = %+v, want a single entry with LocationID %v", got, locationID)
		}
	})

	t.Run("Unknown Farm Maps To Not Found", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{}, dFarm.ErrNotFound
		}}
		h := newTestHandler(userRepo, farmRepo, &fakeLocationRepository{})

		req := httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String()+"/location", nil)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByFarmID(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}

/* ==================== Tests: Update ====================== */

func TestHandler_Update(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	locationID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return dLocation.Location{LocationID: locationID, FarmID: farmID}, nil
			},
			updateLocationFunc: func(context.Context, database.Querier, dLocation.Location) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/location/"+locationID.String(), `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", locationID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.SuccessUpdated.String())
		}
	})

	t.Run("Invalid Location ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/location/not-a-uuid", `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.ErrInvalidID.Error())
		}
	})

	t.Run("Location Belonging To Another Farm Maps To Forbidden", func(t *testing.T) {
		otherFarmID := uuid.New()
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return dLocation.Location{LocationID: locationID, FarmID: otherFarmID}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/location/"+locationID.String(), `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", locationID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusForbidden, rec.Body.String())
		}
	})
}

/* ==================== Tests: Delete ====================== */

func TestHandler_Delete(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	locationID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return dLocation.Location{LocationID: locationID, FarmID: farmID}, nil
			},
			deleteLocationFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/location/"+locationID.String(), nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", locationID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.SuccessDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.SuccessDeleted.String())
		}
	})

	t.Run("Invalid Location ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeLocationRepository{})

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/location/not-a-uuid", nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dLocation.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dLocation.ErrInvalidID.Error())
		}
	})

	t.Run("Unknown Location Maps To Not Found", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return dLocation.Location{}, dLocation.ErrNotFound
			},
		}
		h := newTestHandler(userRepo, farmRepo, locationRepo)

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/location/"+locationID.String(), nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("location_id", locationID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}
