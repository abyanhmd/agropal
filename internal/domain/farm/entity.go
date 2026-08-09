package farm

import (
	"agropal/internal/domain/shared"

	"github.com/google/uuid"
)

type Farm struct {
	ID     uuid.UUID
	Name   string
	UserID uuid.UUID
}

var (
	SuccessCreated shared.SuccessKey = "FARM.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "FARM.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "FARM.DELETE_SUCCESS"
)
