package user

import (
	"github.com/google/uuid"
)

type ModelUser struct {
	UserID   uuid.UUID `db:"user_id"`
	Name     string    `db:"name"`
	Email    string    `db:"email"`
	Username string    `db:"username"`
	Avatar   string    `db:"avatar"`
	Password string    `db:"password"`
	Role     string    `db:"role"`
}
