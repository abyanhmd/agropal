package user

import "github.com/google/uuid"

type User struct {
	UserID   uuid.UUID
	Name     string
	Email    string
	Username string
	Avatar   string
	Password string
	Role     string
}
