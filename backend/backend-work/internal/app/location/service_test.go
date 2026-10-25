package location

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
	dLocation "agropal/internal/domain/location"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
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
// exercised by location.Service, so every other method panics if called - a
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

// fakeFarmRepository implements dFarm.Repository. Only FindFarmByID is
// exercised by location.Service.
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

// fakeLocationRepository implements dLocation.Repository, following the same
// pattern as the fakes above.
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

	t.Run("creates a location for the farm's owner", func(t *testing.T) {
		var created dLocation.Location

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
		locationRepo := &fakeLocationRepository{
			createLocationFunc: func(_ context.Context, _ database.Querier, w dLocation.Location) error {
				created = w
				return nil
			},
		}

		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		_, err := svc.Create(ctxWithUser(ownerID), farmID, LocationCommand{Longitude: 12.34, Latitude: -56.78})
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
		locationRepo := &fakeLocationRepository{
			createLocationFunc: func(context.Context, database.Querier, dLocation.Location) error {
				t.Fatal("CreateLocation should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		_, err := svc.Create(ctxWithUser(otherUserID), farmID, LocationCommand{Longitude: 1, Latitude: 1})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Create() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_GetByFarmID(t *testing.T) {
	ownerID := uuid.New()
	farmID := uuid.New()
	want := []dLocation.Location{
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
	locationRepo := &fakeLocationRepository{
		listLocationsByFarmFunc: func(_ context.Context, _ database.Querier, fid uuid.UUID) ([]dLocation.Location, error) {
			if fid != farmID {
				t.Fatalf("ListLocationsByFarmID called with %v, want %v", fid, farmID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

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
	locationID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	existing := dLocation.Location{LocationID: locationID, FarmID: farmID, Longitude: 1, Latitude: 1}

	newRepos := func(requesterID uuid.UUID, updated *dLocation.Location) (*fakeUserRepository, *fakeFarmRepository, *fakeLocationRepository) {
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
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return existing, nil
			},
			updateLocationFunc: func(_ context.Context, _ database.Querier, w dLocation.Location) error {
				*updated = w
				return nil
			},
		}
		return userRepo, farmRepo, locationRepo
	}

	t.Run("farm owner can update", func(t *testing.T) {
		var updated dLocation.Location
		userRepo, farmRepo, locationRepo := newRepos(ownerID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		err := svc.Update(ctxWithUser(ownerID), farmID, locationID, LocationCommand{Longitude: 9, Latitude: 9})
		if err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Longitude != 9 || updated.Latitude != 9 {
			t.Errorf("updated coords = (%v, %v), want (9, 9)", updated.Longitude, updated.Latitude)
		}
	})

	t.Run("non-owner, non-admin is forbidden", func(t *testing.T) {
		var updated dLocation.Location
		userRepo, farmRepo, locationRepo := newRepos(otherUserID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		err := svc.Update(ctxWithUser(otherUserID), farmID, locationID, LocationCommand{Longitude: 9, Latitude: 9})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Update() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("admin can update any farm's location", func(t *testing.T) {
		var updated dLocation.Location
		userRepo, farmRepo, locationRepo := newRepos(adminID, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		err := svc.Update(ctxWithUserAndRole(adminID, config.RoleAdmin), farmID, locationID, LocationCommand{Longitude: 9, Latitude: 9})
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
	locationID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID}
	existing := dLocation.Location{LocationID: locationID, FarmID: farmID}

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
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return existing, nil
			},
			deleteLocationFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
				deletedID = id
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		if err := svc.Delete(ctxWithUser(ownerID), farmID, locationID); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if deletedID != locationID {
			t.Errorf("deleted location ID = %v, want %v", deletedID, locationID)
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
		locationRepo := &fakeLocationRepository{
			getLocationByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dLocation.Location, error) {
				return existing, nil
			},
			deleteLocationFunc: func(context.Context, database.Querier, uuid.UUID) error {
				t.Fatal("DeleteLocation should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

		err := svc.Delete(ctxWithUser(otherUserID), farmID, locationID)
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
	locationRepo := &fakeLocationRepository{}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo, locationRepo)

	_, err := svc.GetByFarmID(ctxWithUser(uuid.New()), uuid.New())
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("GetByFarmID() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}
