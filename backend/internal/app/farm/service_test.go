package farm

import (
	"agropal/config"
	dFarm "agropal/internal/domain/farm"
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
// exercised by farm.Service, so every other method panics if called -
// a test relying on an un-stubbed call fails loudly instead of silently
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

// fakeFarmRepository implements dFarm.Repository, following the same
// pattern as fakeUserRepository above.
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

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

/* ==================== Tests ====================== */

func TestService_Create(t *testing.T) {
	userID := uuid.New()
	user := dUser.User{ID: userID}

	t.Run("creates a farm owned by the caller", func(t *testing.T) {
		var created dFarm.Farm

		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return user, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			createFarmFunc: func(_ context.Context, _ database.Querier, farm dFarm.Farm) error {
				created = farm
				return nil
			},
		}

		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Create(ctxWithUser(userID), FarmCommand{Name: "Green Acres"})
		if err != nil {
			t.Fatalf("Create() error = %v, want nil", err)
		}
		if created.UserID != userID {
			t.Errorf("created.UserID = %v, want %v", created.UserID, userID)
		}
		if created.Name != "Green Acres" {
			t.Errorf("created.Name = %q, want %q", created.Name, "Green Acres")
		}
		if created.ID == uuid.Nil {
			t.Error("created.ID is zero, want a generated UUID")
		}
	})

	t.Run("propagates the repository error", func(t *testing.T) {
		wantErr := errors.New("insert failed")

		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return user, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			createFarmFunc: func(context.Context, database.Querier, dFarm.Farm) error {
				return wantErr
			},
		}

		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Create(ctxWithUser(userID), FarmCommand{Name: "Green Acres"})
		if !errors.Is(err, wantErr) {
			t.Errorf("Create() error = %v, want %v", err, wantErr)
		}
	})
}

func TestService_GetOne(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	farmID := uuid.New()
	farm := dFarm.Farm{ID: farmID, UserID: ownerID, Name: "Green Acres"}

	t.Run("returns the farm to its owner", func(t *testing.T) {
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

		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		got, err := svc.GetOne(ctxWithUser(ownerID), farmID)
		if err != nil {
			t.Fatalf("GetOne() error = %v, want nil", err)
		}
		if got != farm {
			t.Errorf("GetOne() = %+v, want %+v", got, farm)
		}
	})

	t.Run("forbids a non-owner", func(t *testing.T) {
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

		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		_, err := svc.GetOne(ctxWithUser(otherUserID), farmID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("GetOne() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_Update(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	adminID := uuid.New()
	farmID := uuid.New()
	existing := dFarm.Farm{ID: farmID, UserID: ownerID, Name: "Old Name"}

	newFarmRepo := func(existing dFarm.Farm, updated *dFarm.Farm) *fakeFarmRepository {
		return &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return existing, nil
			},
			updateFarmFunc: func(_ context.Context, _ database.Querier, farm dFarm.Farm) error {
				*updated = farm
				return nil
			},
		}
	}

	t.Run("owner can update their own farm", func(t *testing.T) {
		var updated dFarm.Farm
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := newFarmRepo(existing, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Update(ctxWithUser(ownerID), farmID, FarmCommand{Name: "New Name"})
		if err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("updated.Name = %q, want %q", updated.Name, "New Name")
		}
	})

	t.Run("non-owner, non-admin is forbidden", func(t *testing.T) {
		var updated dFarm.Farm
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherUserID}, nil
			},
		}
		farmRepo := newFarmRepo(existing, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Update(ctxWithUser(otherUserID), farmID, FarmCommand{Name: "New Name"})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Update() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("admin can update any farm", func(t *testing.T) {
		var updated dFarm.Farm
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: adminID}, nil
			},
		}
		farmRepo := newFarmRepo(existing, &updated)
		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Update(ctxWithUserAndRole(adminID, config.RoleAdmin), farmID, FarmCommand{Name: "New Name"})
		if err != nil {
			t.Fatalf("Update() error = %v, want nil", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("updated.Name = %q, want %q", updated.Name, "New Name")
		}
	})
}

func TestService_Delete(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	farmID := uuid.New()
	existing := dFarm.Farm{ID: farmID, UserID: ownerID}

	t.Run("owner can delete their own farm", func(t *testing.T) {
		var deletedID uuid.UUID
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: ownerID}, nil
			},
		}
		farmRepo := &fakeFarmRepository{
			findFarmByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dFarm.Farm, error) {
				return existing, nil
			},
			deleteFarmFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) error {
				deletedID = id
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		if err := svc.Delete(ctxWithUser(ownerID), farmID); err != nil {
			t.Fatalf("Delete() error = %v, want nil", err)
		}
		if deletedID != farmID {
			t.Errorf("deleted farm ID = %v, want %v", deletedID, farmID)
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
				return existing, nil
			},
			deleteFarmFunc: func(context.Context, database.Querier, uuid.UUID) error {
				t.Fatal("DeleteFarm should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, farmRepo)

		err := svc.Delete(ctxWithUser(otherUserID), farmID)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("Delete() error = %v, want %v", err, shared.ErrForbidden)
		}
	})
}

func TestService_GetMany(t *testing.T) {
	userID := uuid.New()
	want := []dFarm.Farm{
		{ID: uuid.New(), UserID: userID, Name: "Farm A"},
		{ID: uuid.New(), UserID: userID, Name: "Farm B"},
	}

	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{ID: userID}, nil
		},
	}
	farmRepo := &fakeFarmRepository{
		listFarmByUserIDFunc: func(_ context.Context, _ database.Querier, uid uuid.UUID) ([]dFarm.Farm, error) {
			if uid != userID {
				t.Fatalf("ListFarmByUserID called with %v, want %v", uid, userID)
			}
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo)

	got, err := svc.GetMany(ctxWithUser(userID))
	if err != nil {
		t.Fatalf("GetMany() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Fatalf("GetMany() returned %d farms, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GetMany()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestService_verifyUser_InvalidUserID(t *testing.T) {
	userRepo := &fakeUserRepository{
		findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
			return dUser.User{}, nil
		},
	}
	farmRepo := &fakeFarmRepository{}
	svc := NewService(fakeTransactor{}, userRepo, farmRepo)

	_, err := svc.GetOne(ctxWithUser(uuid.New()), uuid.New())
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("GetOne() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}
