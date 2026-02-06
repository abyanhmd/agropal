package request

import (
	aUser "agropal/internal/application/user"

	"github.com/google/uuid"
)

type UserUpdateRequest struct {
	Name     *string `json:"name"`
	Username *string `json:"username"`
	Avatar   *string `json:"avatar"`
	Password *string `json:"password"`
	Role     *string `json:"role"`
}

func (r *UserUpdateRequest) ToCommand(userID uuid.UUID) aUser.UserCommand {
	return aUser.UserCommand{
		UserID:   userID,
		Name:     r.Name,
		Username: r.Username,
		Password: r.Password,
	}
}
