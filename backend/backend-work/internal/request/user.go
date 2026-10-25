package request

import (
	aUser "agropal/internal/app/user"

	"github.com/google/uuid"
)

// UserRequest backs auth registration/login, handled by interface/auth's
// own mapper.
type UserRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Username *string `json:"username"`
	Avatar   *string `json:"avatar"`
	Password *string `json:"password"`
}

// UserUpdateRequest backs the profile-update endpoint. Only Name and Avatar
// are currently persisted by the user repository's UPDATE statement;
// Email/Username/Password changes go through dedicated flows (registration
// uniqueness checks, password reset) and are intentionally not accepted
// here.
type UserUpdateRequest struct {
	Name   *string `json:"name"`
	Avatar *string `json:"avatar"`
}

func (r *UserUpdateRequest) ToCommand(userID uuid.UUID) aUser.UserCommand {
	// aUser.UserCommand.ToDomain dereferences every field unconditionally,
	// so fields this endpoint doesn't support editing are filled with
	// harmless placeholders rather than left nil.
	empty := ""

	name := &empty
	if r.Name != nil {
		name = r.Name
	}
	avatar := &empty
	if r.Avatar != nil {
		avatar = r.Avatar
	}

	return aUser.UserCommand{
		UserID:   &userID,
		Name:     name,
		Email:    &empty,
		Username: &empty,
		Avatar:   avatar,
		Password: &empty,
	}
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}
