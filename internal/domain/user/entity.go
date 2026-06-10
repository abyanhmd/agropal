package user

import (
	dAuth "agropal/internal/domain/auth"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Account  AccountInfo
	Security SecurityInfo
	Metadata UserMetadata
}

type AccountInfo struct {
	Name     string
	Email    string
	Username string
	Avatar   string
}

type SecurityInfo struct {
	Password string
	Role     dAuth.Role
}

type UserMetadata struct {
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
	LastLoginAt *time.Time
}
