package user

import (
	"agropal/config"
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID    `json:"id"`
	Account  AccountInfo  `json:"account"`
	Security SecurityInfo `json:"security"`
	Metadata UserMetadata `json:"metadata"`
}

type AccountInfo struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type SecurityInfo struct {
	Password string      `json:"password"`
	Role     config.Role `json:"role"`
}

type UserMetadata struct {
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}

var (
	SuccessUpdated shared.SuccessKey = "USER.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "USER.DELETE_SUCCESS"
)
