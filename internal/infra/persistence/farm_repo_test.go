package persistence

import (
	dFarm "agropal/internal/domain/farm"
	"context"
	"testing"

	"github.com/google/uuid"
)

// insertTestUser inserts a minimal user row directly via SQL (bypassing
// userRepo, which is itself under test elsewhere) so FK-constrained tables
// like farms have a valid user_id to reference.
func insertTestUser(t *testing.T) uuid.UUID {
	t.Helper()
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}

	userID := uuid.New()
	_, err := testDB.Exec(
		`INSERT INTO users (user_id, name, email, username, password, role) VALUES ($1, $2, $3, $4, 'x', 'USER')`,
		userID, "Test User", userID.String()+"@example.com", userID.String(),
	)
	if err != nil {
		t.Fatalf("insertTestUser: %v", err)
	}
	return userID
}

func TestFarmRepository_CreateAndFindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewFarmRepository(testLogger)
	userID := insertTestUser(t)

	farm := dFarm.Farm{ID: uuid.New(), Name: "Green Acres", UserID: userID}
	if err := repo.CreateFarm(ctx, testDB, farm); err != nil {
		t.Fatalf("CreateFarm() error = %v, want nil", err)
	}

	got, err := repo.FindFarmByID(ctx, testDB, farm.ID)
	if err != nil {
		t.Fatalf("FindFarmByID() error = %v, want nil", err)
	}
	if got != farm {
		t.Errorf("FindFarmByID() = %+v, want %+v", got, farm)
	}
}

func TestFarmRepository_FindFarmByID_NotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewFarmRepository(testLogger)

	_, err := repo.FindFarmByID(ctx, testDB, uuid.New())
	if err == nil {
		t.Fatal("FindFarmByID() error = nil, want a not-found error for a nonexistent farm")
	}
}

func TestFarmRepository_ListFarmByUserID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewFarmRepository(testLogger)
	userID := insertTestUser(t)
	otherUserID := insertTestUser(t)

	farmA := dFarm.Farm{ID: uuid.New(), Name: "Farm A", UserID: userID}
	farmB := dFarm.Farm{ID: uuid.New(), Name: "Farm B", UserID: userID}
	farmOther := dFarm.Farm{ID: uuid.New(), Name: "Someone Else's Farm", UserID: otherUserID}
	for _, f := range []dFarm.Farm{farmA, farmB, farmOther} {
		if err := repo.CreateFarm(ctx, testDB, f); err != nil {
			t.Fatalf("CreateFarm(%q) error = %v, want nil", f.Name, err)
		}
	}

	got, err := repo.ListFarmByUserID(ctx, testDB, userID)
	if err != nil {
		t.Fatalf("ListFarmByUserID() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListFarmByUserID() returned %d farms, want 2 (not the other user's farm)", len(got))
	}
	for _, f := range got {
		if f.UserID != userID {
			t.Errorf("ListFarmByUserID() returned a farm belonging to %v, want only %v", f.UserID, userID)
		}
	}
}

func TestFarmRepository_UpdateFarm(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewFarmRepository(testLogger)
	userID := insertTestUser(t)

	farm := dFarm.Farm{ID: uuid.New(), Name: "Old Name", UserID: userID}
	if err := repo.CreateFarm(ctx, testDB, farm); err != nil {
		t.Fatalf("CreateFarm() error = %v, want nil", err)
	}

	farm.Name = "New Name"
	if err := repo.UpdateFarm(ctx, testDB, farm); err != nil {
		t.Fatalf("UpdateFarm() error = %v, want nil", err)
	}

	got, err := repo.FindFarmByID(ctx, testDB, farm.ID)
	if err != nil {
		t.Fatalf("FindFarmByID() error = %v, want nil", err)
	}
	if got.Name != "New Name" {
		t.Errorf("FindFarmByID().Name = %q, want %q", got.Name, "New Name")
	}
}

func TestFarmRepository_DeleteFarm(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewFarmRepository(testLogger)
	userID := insertTestUser(t)

	farm := dFarm.Farm{ID: uuid.New(), Name: "To Be Deleted", UserID: userID}
	if err := repo.CreateFarm(ctx, testDB, farm); err != nil {
		t.Fatalf("CreateFarm() error = %v, want nil", err)
	}

	if err := repo.DeleteFarm(ctx, testDB, farm.ID); err != nil {
		t.Fatalf("DeleteFarm() error = %v, want nil", err)
	}

	if _, err := repo.FindFarmByID(ctx, testDB, farm.ID); err == nil {
		t.Error("FindFarmByID() error = nil after DeleteFarm, want a not-found error")
	}
}
