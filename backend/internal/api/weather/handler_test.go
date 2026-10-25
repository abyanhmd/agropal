package weather

import (
	"agropal/config"
	aWeather "agropal/internal/app/weather"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
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
// exercised by the weather service; everything else panics if called.
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
func (f *fakeUserRepository) ListUsersByAsset(context.Context, database.Querier, shared.AssetType, uuid.UUID) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsersByAsset not stubbed")
}

// fakeFarmRepository implements dFarm.Repository. Only FindFarmByID is
// exercised by the weather service.
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

// fakeWeatherRepository implements dWeather.Repository.
type fakeWeatherRepository struct {
	createWeatherLocationFunc      func(ctx context.Context, q database.Querier, weather dWeather.Weather) error
	getWeatherLocationByIDFunc     func(ctx context.Context, q database.Querier, id uuid.UUID) (dWeather.Weather, error)
	listWeatherLocationsByFarmFunc func(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dWeather.Weather, error)
	updateWeatherLocationFunc      func(ctx context.Context, q database.Querier, weather dWeather.Weather) error
	deleteWeatherLocationFunc      func(ctx context.Context, q database.Querier, id uuid.UUID) error
}

func (f *fakeWeatherRepository) CreateWeatherLocation(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	if f.createWeatherLocationFunc == nil {
		panic("fakeWeatherRepository: createWeatherLocationFunc not stubbed")
	}
	return f.createWeatherLocationFunc(ctx, q, weather)
}
func (f *fakeWeatherRepository) GetWeatherLocationByID(ctx context.Context, q database.Querier, id uuid.UUID) (dWeather.Weather, error) {
	if f.getWeatherLocationByIDFunc == nil {
		panic("fakeWeatherRepository: getWeatherLocationByIDFunc not stubbed")
	}
	return f.getWeatherLocationByIDFunc(ctx, q, id)
}
func (f *fakeWeatherRepository) ListWeatherLocationsByFarmID(ctx context.Context, q database.Querier, farmID uuid.UUID) ([]dWeather.Weather, error) {
	if f.listWeatherLocationsByFarmFunc == nil {
		panic("fakeWeatherRepository: listWeatherLocationsByFarmFunc not stubbed")
	}
	return f.listWeatherLocationsByFarmFunc(ctx, q, farmID)
}
func (f *fakeWeatherRepository) UpdateWeatherLocation(ctx context.Context, q database.Querier, weather dWeather.Weather) error {
	if f.updateWeatherLocationFunc == nil {
		panic("fakeWeatherRepository: updateWeatherLocationFunc not stubbed")
	}
	return f.updateWeatherLocationFunc(ctx, q, weather)
}
func (f *fakeWeatherRepository) DeleteWeatherLocation(ctx context.Context, q database.Querier, id uuid.UUID) error {
	if f.deleteWeatherLocationFunc == nil {
		panic("fakeWeatherRepository: deleteWeatherLocationFunc not stubbed")
	}
	return f.deleteWeatherLocationFunc(ctx, q, id)
}

/* ==================== Helpers ====================== */

func newTestHandler(userRepo dUser.Repository, farmRepo dFarm.Repository, weatherRepo dWeather.Repository) Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aWeather.NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)
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
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/weather", `{"longitude":10.5,"latitude":20.5}`)
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
		weatherRepo := &fakeWeatherRepository{createWeatherLocationFunc: func(context.Context, database.Querier, dWeather.Weather) error {
			return nil
		}}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/weather", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.SuccessCreated.String())
		}
	})

	t.Run("Invalid Farm ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		req := jsonRequest(http.MethodPost, "/farm/not-a-uuid/weather", `{"longitude":10.5,"latitude":20.5}`)
		req.SetPathValue("farm_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.ErrInvalidFarmID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.ErrInvalidFarmID.Error())
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/weather", "not-json")
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
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		// (0, 0) is a valid coordinate, so the "longitude"/"latitude" tags
		// only reject values outside their real-world range.
		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/weather", `{"longitude":200,"latitude":200}`)
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
		h := newTestHandler(userRepo, farmRepo, &fakeWeatherRepository{})

		req := jsonRequest(http.MethodPost, "/farm/"+farmID.String()+"/weather", `{"longitude":10.5,"latitude":20.5}`)
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
		weatherID := uuid.New()
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		weatherRepo := &fakeWeatherRepository{listWeatherLocationsByFarmFunc: func(context.Context, database.Querier, uuid.UUID) ([]dWeather.Weather, error) {
			return []dWeather.Weather{{LocationID: weatherID, FarmID: farmID, Longitude: 1, Latitude: 2}}, nil
		}}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String()+"/weather", nil)
		req.SetPathValue("farm_id", farmID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.GetByFarmID(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []response.WeatherResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if len(got) != 1 || got[0].WeatherID != weatherID {
			t.Errorf("got = %+v, want a single entry with WeatherID %v", got, weatherID)
		}
	})

	t.Run("Unknown Farm Maps To Not Found", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{}, dFarm.ErrNotFound
		}}
		h := newTestHandler(userRepo, farmRepo, &fakeWeatherRepository{})

		req := httptest.NewRequest(http.MethodGet, "/farm/"+farmID.String()+"/weather", nil)
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
	weatherID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return dWeather.Weather{LocationID: weatherID, FarmID: farmID}, nil
			},
			updateWeatherLocationFunc: func(context.Context, database.Querier, dWeather.Weather) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/weather/"+weatherID.String(), `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", weatherID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.SuccessUpdated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.SuccessUpdated.String())
		}
	})

	t.Run("Invalid Weather ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/weather/not-a-uuid", `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Update(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.ErrInvalidID.Error())
		}
	})

	t.Run("Weather Location Belonging To Another Farm Maps To Forbidden", func(t *testing.T) {
		otherFarmID := uuid.New()
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return dWeather.Weather{LocationID: weatherID, FarmID: otherFarmID}, nil
			},
		}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := jsonRequest(http.MethodPatch, "/farm/"+farmID.String()+"/weather/"+weatherID.String(), `{"longitude":11,"latitude":22}`)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", weatherID.String())
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
	weatherID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return dWeather.Weather{LocationID: weatherID, FarmID: farmID}, nil
			},
			deleteWeatherLocationFunc: func(context.Context, database.Querier, uuid.UUID) error { return nil },
		}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/weather/"+weatherID.String(), nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", weatherID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.SuccessDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.SuccessDeleted.String())
		}
	})

	t.Run("Invalid Weather ID Path Value", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakeFarmRepository{}, &fakeWeatherRepository{})

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/weather/not-a-uuid", nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dWeather.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dWeather.ErrInvalidID.Error())
		}
	})

	t.Run("Unknown Weather Location Maps To Not Found", func(t *testing.T) {
		userRepo := &fakeUserRepository{findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		}}
		farmRepo := &fakeFarmRepository{findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		}}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return dWeather.Weather{}, dWeather.ErrNotFound
			},
		}
		h := newTestHandler(userRepo, farmRepo, weatherRepo)

		req := httptest.NewRequest(http.MethodDelete, "/farm/"+farmID.String()+"/weather/"+weatherID.String(), nil)
		req.SetPathValue("farm_id", farmID.String())
		req.SetPathValue("weather_id", weatherID.String())
		req = authedRequest(req, ownerID, config.RoleUser)
		rec := httptest.NewRecorder()

		h.Delete(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}
