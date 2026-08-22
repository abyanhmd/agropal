package persistence

import (
	dWeather "agropal/internal/domain/weather"
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

func TestWeatherRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewWeatherRepository(testLogger)
	farmID := insertTestFarm(t)

	weather := dWeather.Weather{LocationID: uuid.New(), Longitude: 12.34, Latitude: -56.78, FarmID: farmID}
	if err := repo.CreateWeatherLocation(ctx, testDB, weather); err != nil {
		t.Fatalf("CreateWeatherLocation() error = %v, want nil", err)
	}

	got, err := repo.GetWeatherLocationByID(ctx, testDB, weather.LocationID)
	if err != nil {
		t.Fatalf("GetWeatherLocationByID() error = %v, want nil", err)
	}
	if got != weather {
		t.Errorf("GetWeatherLocationByID() = %+v, want %+v", got, weather)
	}
}

func TestWeatherRepository_GetByID_NotFound(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewWeatherRepository(testLogger)

	_, err := repo.GetWeatherLocationByID(ctx, testDB, uuid.New())
	if err == nil {
		t.Fatal("GetWeatherLocationByID() error = nil, want a not-found error for a nonexistent location")
	}
}

func TestWeatherRepository_ListWeatherLocationsByFarmID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewWeatherRepository(testLogger)
	farmID := insertTestFarm(t)
	otherFarmID := insertTestFarm(t)

	locA := dWeather.Weather{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	locB := dWeather.Weather{LocationID: uuid.New(), Longitude: 2, Latitude: 2, FarmID: farmID}
	locOther := dWeather.Weather{LocationID: uuid.New(), Longitude: 3, Latitude: 3, FarmID: otherFarmID}
	for _, loc := range []dWeather.Weather{locA, locB, locOther} {
		if err := repo.CreateWeatherLocation(ctx, testDB, loc); err != nil {
			t.Fatalf("CreateWeatherLocation() error = %v, want nil", err)
		}
	}

	got, err := repo.ListWeatherLocationsByFarmID(ctx, testDB, farmID)
	if err != nil {
		t.Fatalf("ListWeatherLocationsByFarmID() error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Fatalf("ListWeatherLocationsByFarmID() returned %d locations, want 2 (not the other farm's)", len(got))
	}
	for _, loc := range got {
		if loc.FarmID != farmID {
			t.Errorf("ListWeatherLocationsByFarmID() returned a location belonging to %v, want only %v", loc.FarmID, farmID)
		}
	}
}

func TestWeatherRepository_UpdateWeatherLocation(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewWeatherRepository(testLogger)
	farmID := insertTestFarm(t)

	weather := dWeather.Weather{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	if err := repo.CreateWeatherLocation(ctx, testDB, weather); err != nil {
		t.Fatalf("CreateWeatherLocation() error = %v, want nil", err)
	}

	weather.Longitude = 9.9
	weather.Latitude = -9.9
	if err := repo.UpdateWeatherLocation(ctx, testDB, weather); err != nil {
		t.Fatalf("UpdateWeatherLocation() error = %v, want nil", err)
	}

	got, err := repo.GetWeatherLocationByID(ctx, testDB, weather.LocationID)
	if err != nil {
		t.Fatalf("GetWeatherLocationByID() error = %v, want nil", err)
	}
	if got.Longitude != 9.9 || got.Latitude != -9.9 {
		t.Errorf("updated coords = (%v, %v), want (9.9, -9.9)", got.Longitude, got.Latitude)
	}
}

func TestWeatherRepository_DeleteWeatherLocation(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewWeatherRepository(testLogger)
	farmID := insertTestFarm(t)

	weather := dWeather.Weather{LocationID: uuid.New(), Longitude: 1, Latitude: 1, FarmID: farmID}
	if err := repo.CreateWeatherLocation(ctx, testDB, weather); err != nil {
		t.Fatalf("CreateWeatherLocation() error = %v, want nil", err)
	}

	if err := repo.DeleteWeatherLocation(ctx, testDB, weather.LocationID); err != nil {
		t.Fatalf("DeleteWeatherLocation() error = %v, want nil", err)
	}

	if _, err := repo.GetWeatherLocationByID(ctx, testDB, weather.LocationID); err == nil {
		t.Error("GetWeatherLocationByID() error = nil after DeleteWeatherLocation, want a not-found error")
	}
}
