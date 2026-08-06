package auth

import (
	dAuth "agropal/internal/domain/auth"
	dUser "agropal/internal/domain/user"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type CreateUserCommand struct {
	Name     string
	Email    string
	Username string
	Password string
	Avatar   *string
}

type LoginUserCommand struct {
	Email    *string
	Username *string
	Password string
}

type ChangePasswordCommand struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
}

/* ==================== Mapping ====================== */

func (c CreateUserCommand) ToDomain(userID uuid.UUID) dUser.User {
	return dUser.User{
		ID: userID,
		Account: dUser.AccountInfo{
			Name:     c.Name,
			Email:    c.Email,
			Username: c.Username,
			Avatar:   *c.Avatar,
		},
		Security: dUser.SecurityInfo{
			Password: c.Password,
			Role:     dAuth.RoleUser,
		},
	}
}
