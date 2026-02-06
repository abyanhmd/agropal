package user

import (
	dNote "agropal/internal/domain/note"
	"agropal/internal/domain/shared"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user User) error
	GetOne(ctx context.Context, userID uuid.UUID) (User, error)
	Update(ctx context.Context, user User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, password string) error

	SoftDelete(ctx context.Context, deleteNote dNote.Note) error
	HardDelete(ctx context.Context, userID uuid.UUID, deleteNote dNote.Note) error

	GetByAdmin(ctx context.Context, assetType shared.AssetType, assetID uuid.UUID) ([]User, error)
}
