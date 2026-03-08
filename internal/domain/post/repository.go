package post

import (
	"agropal/internal/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, q database.Querier, p Post) error
	Update(ctx context.Context, q database.Querier, p Post) error
	Delete(ctx context.Context, q database.Querier, id uuid.UUID) error

	FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (Post, error)
	FindBySlug(ctx context.Context, q database.Querier, slug string) (Post, error)
	List(ctx context.Context, q database.Querier, limit, offset int) ([]Post, error)

	AddComment(ctx context.Context, q database.Querier, c PostComment) error
	ListTopLevel(ctx context.Context, q database.Querier, postID uuid.UUID, limit, offset int) ([]PostComment, error)
	ListReplies(ctx context.Context, q database.Querier, parentID uuid.UUID, limit, offset int) ([]PostComment, error)

	AddEngagement(ctx context.Context, q database.Querier, e PostEngagement) error
	SyncCounts(ctx context.Context, q database.Querier, assetType PostType, assetID uuid.UUID) error
}
