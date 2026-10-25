package response

import "github.com/google/uuid"

type FarmResponse struct {
	FarmID uuid.UUID `json:"farm_id"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}
