package response

import "github.com/google/uuid"

type UserResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
	Password string    `json:"password"`
	Role     string    `json:"role"`
}
