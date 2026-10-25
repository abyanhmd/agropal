package location

import (
	dLocation "agropal/internal/domain/location"
	"testing"

	"github.com/google/uuid"
)

func TestLocationCommand_ToDomain(t *testing.T) {
	locationID := uuid.New()
	farmID := uuid.New()
	cmd := LocationCommand{Longitude: 12.34, Latitude: -56.78}

	got := cmd.ToDomain(locationID, farmID)

	want := dLocation.Location{
		LocationID: locationID,
		Longitude:  12.34,
		Latitude:   -56.78,
		FarmID:     farmID,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
