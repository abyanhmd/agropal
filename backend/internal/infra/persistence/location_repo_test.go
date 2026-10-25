package persistence

import (
	dLocation "agropal/internal/domain/location"
	"context"
	"testing"

	"github.com/google/uuid"
)

// insertTestFarm inserts a farm directly via SQL (bypassing farmRepo, which
// is itself under test elsewhere) so FK-constrained tables like
// location_info have a valid farm_id to reference.
func insertTestFarm(t *testing.T) uuid.UUID {
	t.Helper()
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}

	userID := insertTestUser(t)
	farmID := uuid.New()
	_, err := testDB.Exec(
		`INSERT INTO farms (farm_id, user_id, name) VALUES ($1, $2, $3)`,
		farmID, userID, "Test Farm",
	)
	if err != nil {
		t.Fatalf("insertTestFarm: %v", err)
	}
	return farmID
}

func TestLocationRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewLocationRepository(testLogger)
	farmID := insertTestFarm(t)

	location := dLocation.Location{LocationID: uuid.New(), Longitude: 12.34, Latitude: -56.78, FarmID: farmID}
	if err := repo.CreateLocation(ctx, testDB, location); err != nil {
		t.Fatalf("CreateLocation() error = %v, want nil", err)
	}

	got, err := repo.GetLocationByID(ctx, testDB, location.LocationID)
	if err != nil {
		t.Fatalf("GetLocationByID() error = %v, want nil", err)
	}
	if got != location {
		t.Errorf("GetLocationByID() = %+v, want %+v", got, location)
	}
}

func TestLocationRepository_GetByID_NotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewLocationRepository(testLogger)

	_, err := repo.GetLocationByID(ctx, testDB, uuid.New())
	if err == nil {
		t.Fatal("GetLocationByID() error = nil, want a not-found error for a nonexistent location")
	}
}

func TestLocationRepository_ListLocationsByFarmID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewLocationRepository(testLogger)
	farmID := insertTestFarm(t)
	otherFarmID := insertTestFarm(t)

	locA := dLocation.Location{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	locB := dLocation.Location{LocationID: uuid.New(), Longitude: 2, Latitude: 2, FarmID: farmID}
	locOther := dLocation.Location{LocationID: uuid.New(), Longitude: 3, Latitude: 3, FarmID: otherFarmID}
	for _, loc := range []dLocation.Location{locA, locB, locOther} {
		if err := repo.CreateLocation(ctx, testDB, loc); err != nil {
			t.Fatalf("CreateLocation() error = %v, want nil", err)
		}
	}

	got, err := repo.ListLocationsByFarmID(ctx, testDB, farmID)
	if err != nil {
		t.Fatalf("ListLocationsByFarmID() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListLocationsByFarmID() returned %d locations, want 2 (not the other farm's)", len(got))
	}
	for _, loc := range got {
		if loc.FarmID != farmID {
			t.Errorf("ListLocationsByFarmID() returned a location belonging to %v, want only %v", loc.FarmID, farmID)
		}
	}
}

func TestLocationRepository_UpdateLocation(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewLocationRepository(testLogger)
	farmID := insertTestFarm(t)

	location := dLocation.Location{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	if err := repo.CreateLocation(ctx, testDB, location); err != nil {
		t.Fatalf("CreateLocation() error = %v, want nil", err)
	}

	location.Longitude = 9.9
	location.Latitude = -9.9
	if err := repo.UpdateLocation(ctx, testDB, location); err != nil {
		t.Fatalf("UpdateLocation() error = %v, want nil", err)
	}

	got, err := repo.GetLocationByID(ctx, testDB, location.LocationID)
	if err != nil {
		t.Fatalf("GetLocationByID() error = %v, want nil", err)
	}
	if got.Longitude != 9.9 || got.Latitude != -9.9 {
		t.Errorf("updated coords = (%v, %v), want (9.9, -9.9)", got.Longitude, got.Latitude)
	}
}

func TestLocationRepository_DeleteLocation(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewLocationRepository(testLogger)
	farmID := insertTestFarm(t)

	location := dLocation.Location{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	if err := repo.CreateLocation(ctx, testDB, location); err != nil {
		t.Fatalf("CreateLocation() error = %v, want nil", err)
	}

	if err := repo.DeleteLocation(ctx, testDB, location.LocationID); err != nil {
		t.Fatalf("DeleteLocation() error = %v, want nil", err)
	}

	if _, err := repo.GetLocationByID(ctx, testDB, location.LocationID); err == nil {
		t.Error("GetLocationByID() error = nil after DeleteLocation, want a not-found error")
	}
}
