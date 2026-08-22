package auth

import (
	"agropal/config"
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
	var avatar string
	if c.Avatar != nil {
		avatar = *c.Avatar
	}

	return dUser.User{
		ID: userID,
		Account: dUser.AccountInfo{
			Name:     c.Name,
			Email:    c.Email,
			Username: c.Username,
			Avatar:   avatar,
		},
		Security: dUser.SecurityInfo{
			Password: c.Password,
			Role:     config.RoleUser,
		},
	}
}
