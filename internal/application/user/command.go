package user

import (
	dUser "agropal/internal/domain/user"

	"github.com/google/uuid"
)

/* ==================== Struct ====================== */

type UserCommand struct {
	UserID      uuid.UUID `json:"user_id"`
	Name        *string   `json:"name" validate:"required"`
	Email       *string   `json:"email" validate:"required,email"`
	Username    *string   `json:"username" validate:"required,alphanum"`
	PhoneNumber *string   `json:"phone_number" validate:"e164"`
	Password    *string   `json:"password" validate:"required,min=8"`
}

type LoginCommand struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordCommand struct {
	Email       string `json:"email" validate:"required,email"`
	OldPassword string `json:"password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

/* ==================== Mapping ====================== */

func (c UserCommand) ToDomain() dUser.User {
	return dUser.User{
		Name:     *c.Name,
		Email:    *c.Email,
		Username: *c.Username,
		Password: *c.Password,
	}
}
