package auth

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"testing"

	"github.com/google/uuid"
)

func TestCreateUserCommand_ToDomain(t *testing.T) {
	userID := uuid.New()

	cmd := CreateUserCommand{
		Name:     "Jane Farmer",
		Email:    "jane@example.com",
		Username: "janefarmer",
		Password: "hashed-password",
		Avatar:   shared.Ptr("avatar.png"),
	}

	got := cmd.ToDomain(userID)

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
			Role:     config.RoleUser,
		},
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}

// Avatar is optional on registration (internal/request.UserRequest.Avatar
// has no `validate:"required"` tag), so ToDomain must tolerate a nil
// pointer instead of dereferencing it directly.
func TestCreateUserCommand_ToDomain_NilAvatar(t *testing.T) {
	userID := uuid.New()

	cmd := CreateUserCommand{
		Name:     "Jane Farmer",
		Email:    "jane@example.com",
		Username: "janefarmer",
		Password: "hashed-password",
		Avatar:   nil,
	}

	got := cmd.ToDomain(userID)

	if got.Account.Avatar != "" {
		t.Errorf("Account.Avatar = %q, want empty string for nil Avatar", got.Account.Avatar)
	}
}
