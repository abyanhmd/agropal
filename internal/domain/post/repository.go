package post

import (
	"agropal/internal/infra/database"
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePost(ctx context.Context, q database.Querier, p Post) error
	UpdatePost(ctx context.Context, q database.Querier, p Post) error
	DeletePost(ctx context.Context, q database.Querier, id uuid.UUID) error

	FindPostByID(ctx context.Context, q database.Querier, id uuid.UUID) (Post, error)
	FindPostBySlug(ctx context.Context, q database.Querier, slug string) (Post, error)
	ListPosts(ctx context.Context, q database.Querier, filter FilterQuery) ([]Post, error)

	AddComment(ctx context.Context, q database.Querier, c PostComment) error
	ListTopLevelComments(ctx context.Context, q database.Querier, postID uuid.UUID, filter FilterQuery) ([]PostComment, error)
	ListCommentReplies(ctx context.Context, q database.Querier, parentID uuid.UUID, filter FilterQuery) ([]PostComment, error)

	AddEngagement(ctx context.Context, q database.Querier, e PostEngagement) error
	SyncPostEngagements(ctx context.Context, q database.Querier, assetType PostType, assetID uuid.UUID) error
}
