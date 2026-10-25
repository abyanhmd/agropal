package user

import (
	dUser "agropal/internal/domain/user"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type UserCommand struct {
	UserID   *uuid.UUID `json:"user_id"`
	Name     *string    `json:"name" validate:"required"`
	Email    *string    `json:"email" validate:"required,email"`
	Username *string    `json:"username" validate:"required,alphanum"`
	Avatar   *string    `json:"avatar"`
	Password *string    `json:"password" validate:"required,min=8"`
}

type UpdateUserCommand struct {
	Email    *string
	Username *string
	Name     *string
	Avatar   *string
}

/* ==================== Mapping ====================== */

func (c UserCommand) ToDomain() dUser.User {
	return dUser.User{
		ID: *c.UserID,
		Account: dUser.AccountInfo{
			Name:     *c.Name,
			Email:    *c.Email,
			Username: *c.Username,
			Avatar:   *c.Avatar,
		},
		Security: dUser.SecurityInfo{
			Password: *c.Password,
		},
	}
}
