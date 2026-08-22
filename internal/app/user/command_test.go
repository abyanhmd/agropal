package user

import (
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"testing"

	"github.com/google/uuid"
)

func TestUserCommand_ToDomain(t *testing.T) {
	userID := uuid.New()

	cmd := UserCommand{
		UserID:   &userID,
		Name:     shared.Ptr("Jane Farmer"),
		Email:    shared.Ptr("jane@example.com"),
		Username: shared.Ptr("janefarmer"),
		Avatar:   shared.Ptr("avatar.png"),
		Password: shared.Ptr("hashed-password"),
	}

	got := cmd.ToDomain()

	want := dUser.User{
		ID: userID,
		Account: dUser.AccountInfo{
			Name:     "Jane Farmer",
			Email:    "jane@example.com",
			Username: "janefarmer",
			Avatar:   "avatar.png",
		},
		Security: dUser.SecurityInfo{
			Password: "hashed-password",
		},
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
