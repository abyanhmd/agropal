package user

import (
	"time"

	"github.com/google/uuid"
)

type UserModel struct {
	ID       uuid.UUID `db:"user_id"`
	Name     string    `db:"name"`
	Email    string    `db:"email"`
	Username string    `db:"username"`
	Avatar   string    `db:"avatar"`
	Password string    `db:"password"`
	Role     string    `db:"role"`

	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
	LastLoginAt *time.Time `db:"last_login_at"`
}
