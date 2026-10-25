package farm

import (
	dFarm "agropal/internal/domain/farm"
	"testing"

	"github.com/google/uuid"
)

func TestFarmCommand_ToDomain(t *testing.T) {
	userID := uuid.New()
	cmd := FarmCommand{Name: "Green Acres"}

	got := cmd.ToDomain(userID)

	want := dFarm.Farm{
		UserID: userID,
		Name:   "Green Acres",
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
