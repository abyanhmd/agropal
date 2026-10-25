package weather

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/infra/database"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier
// instead of opening a real transaction. Safe here because the fake
// repositories below never touch the Querier they're given - they stand in
// for the whole repository, not just the SQL layer underneath it.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier {
	return nil
}

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by weather.Service, so every other method panics if called - a
// test relying on an un-stubbed call fails loudly instead of silently
// zero-valuing.
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
// exercised by weather.Service.
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

// fakeWeatherRepository implements dWeather.Repository, following the same
// pattern as the fakes above.
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

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

/* ==================== Tests ====================== */

func TestService_Create(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}

	t.Run("creates a weather location for the farm's owner", func(t *testing.T) {
		var created dWeather.Weather

		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		weatherRepo := &fakeWeatherRepository{
			createWeatherLocationFunc: func(_ context.Context, _ database.Querier, w dWeather.Weather) error {
				created = w
				return nil
			},
		}

		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Create(ctxWithUser(ownerID), farmID, WeatherCommand{Longitude: 12.34, Latitude: -56.78})
		if err != nil {
			t.Fatalf("Create() error = %v, want nil", err)
		}
		if created.FarmID != farmID {
			t.Errorf("created.FarmID = %v, want %v", created.FarmID, farmID)
		}
		if created.Longitude != 12.34 || created.Latitude != -56.78 {
			t.Errorf("created coords = (%v, %v), want (12.34, -56.78)", created.Longitude, created.Latitude)
		}
	})

	t.Run("forbids a non-owner", func(t *testing.T) {
		otherUserID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		weatherRepo := &fakeWeatherRepository{
			createWeatherLocationFunc: func(context.Context, database.Querier, dWeather.Weather) error {
				t.Fatal("CreateWeatherLocation should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Create(ctxWithUser(otherUserID), farmID, WeatherCommand{Longitude: 1, Latitude: 1})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Create() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_GetByFarmID(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	want := []dWeather.Weather{
		{LocationID: uuid.New(), FarmID: farmID, Longitude: 1, Latitude: 2},
		{LocationID: uuid.New(), FarmID: farmID, Longitude: 3, Latitude: 4},
	}

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: ownerID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
			return dFarm.Farm{ID: farmID, UserID: ownerID}, nil
		},
	}
	weatherRepo := &fakeWeatherRepository{
		listWeatherLocationsByFarmFunc: func(_ context.Context, _ database.Querier, fid uuid.UUID) ([]dWeather.Weather, error) {
			if fid != farmID {
				t.Fatalf("ListWeatherLocationsByFarmID called with %v, want %v", fid, farmID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

	got, err := svc.GetByFarmID(ctxWithUser(ownerID), farmID)
	if err != nil {
		t.Fatalf("GetByFarmID() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Fatalf("GetByFarmID() returned %d locations, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetByFarmID()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestService_Update(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	adminID := uuid.New()
	farmID := uuid.New()
	weatherID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	existing := dWeather.Weather{LocationID: weatherID, FarmID: farmID, Longitude: 1, Latitude: 1}

	newRepos := func(requesterID uuid.UUID, updated *dWeather.Weather) (*fakeUserRepository, *fakeFarmRepository, *fakeWeatherRepository) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: requesterID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return existing, nil
			},
			updateWeatherLocationFunc: func(_ context.Context, _ database.Querier, w dWeather.Weather) error {
				*updated = w
				return nil
			},
		}
		return userRepo, farmRepo, weatherRepo
	}

	t.Run("farm owner can update", func(t *testing.T) {
		var updated dWeather.Weather
		userRepo, farmRepo, weatherRepo := newRepos(ownerID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Update(ctxWithUser(ownerID), farmID, weatherID, WeatherCommand{Longitude: 9, Latitude: 9})
		if err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Longitude != 9 || updated.Latitude != 9 {
			t.Errorf("updated coords = (%v, %v), want (9, 9)", updated.Longitude, updated.Latitude)
		}
	})

	t.Run("non-owner, non-admin is forbidden", func(t *testing.T) {
		var updated dWeather.Weather
		userRepo, farmRepo, weatherRepo := newRepos(otherUserID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Update(ctxWithUser(otherUserID), farmID, weatherID, WeatherCommand{Longitude: 9, Latitude: 9})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Update() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("admin can update any farm's weather location", func(t *testing.T) {
		var updated dWeather.Weather
		userRepo, farmRepo, weatherRepo := newRepos(adminID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Update(ctxWithUserAndRole(adminID, config.RoleAdmin), farmID, weatherID, WeatherCommand{Longitude: 9, Latitude: 9})
		if err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Longitude != 9 || updated.Latitude != 9 {
			t.Errorf("updated coords = (%v, %v), want (9, 9)", updated.Longitude, updated.Latitude)
		}
	})
}

func TestService_Delete(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	farmID := uuid.New()
	weatherID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	existing := dWeather.Weather{LocationID: weatherID, FarmID: farmID}

	t.Run("farm owner can delete", func(t *testing.T) {
		var deletedID uuid.UUID
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return existing, nil
			},
			deleteWeatherLocationFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
				deletedID = id
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		if err := svc.Delete(ctxWithUser(ownerID), farmID, weatherID); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if deletedID != weatherID {
			t.Errorf("deleted weather ID = %v, want %v", deletedID, weatherID)
		}
	})

	t.Run("non-owner, non-admin is forbidden", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return farm, nil
			},
		}
		weatherRepo := &fakeWeatherRepository{
			getWeatherLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dWeather.Weather, error) {
				return existing, nil
			},
			deleteWeatherLocationFunc: func(context.Context, database.Querier, uuid.UUID) error {
				t.Fatal("DeleteWeatherLocation should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

		err := svc.Delete(ctxWithUser(otherUserID), farmID, weatherID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Delete() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_verifyFarm_InvalidUserID(t *testing.T) {
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{}, nil
		},
	}
	farmRepo := &fakeFarmRepository{}
	weatherRepo := &fakeWeatherRepository{}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, weatherRepo)

	_, err := svc.GetByFarmID(ctxWithUser(uuid.New()), uuid.New())
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("GetByFarmID() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}
