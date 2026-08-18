package user

import (
	"agropal/config"
	"agropal/internal/domain/shared"
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
	Role     config.Role
}

type UserMetadata struct {
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	DeletedAt       *time.Time
	LastLoginAt     *time.Time
	EmailVerifiedAt *time.Time
}

var (
	SuccessUpdated shared.SuccessKey = "USER.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "USER.DELETE_SUCCESS"
)
