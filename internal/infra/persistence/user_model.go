package persistence

import (
	dAuth "agropal/internal/domain/auth"
	dUser "agropal/internal/domain/user"
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

func ToUserModel(domain dUser.User) UserModel {
	return UserModel{
		ID: domain.ID,

		Name:     domain.Account.Name,
		Email:    domain.Account.Email,
		Username: domain.Account.Username,
		Avatar:   domain.Account.Avatar,

		Password: domain.Security.Password,
		Role:     string(domain.Security.Role),

		CreatedAt:   domain.Metadata.CreatedAt,
		UpdatedAt:   domain.Metadata.UpdatedAt,
		DeletedAt:   domain.Metadata.DeletedAt,
		LastLoginAt: domain.Metadata.LastLoginAt,
	}
}

func (m UserModel) ToDomain() dUser.User {
	return dUser.User{
		ID: m.ID,
		Account: dUser.AccountInfo{
			Name:     m.Name,
			Email:    m.Email,
			Username: m.Username,
			Avatar:   m.Avatar,
		},
		Security: dUser.SecurityInfo{
			Password: m.Password,
			Role:     dAuth.Role(m.Role),
		},
		Metadata: dUser.UserMetadata{
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			DeletedAt:   m.DeletedAt,
			LastLoginAt: m.LastLoginAt,
		},
	}
}
