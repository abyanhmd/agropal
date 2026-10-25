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
	// CountPosts counts the posts a filter matches, ignoring its pagination.
	CountPosts(ctx context.Context, q database.Querier, filter FilterQuery) (int, error)
	ListPostAuthors(ctx context.Context, q database.Querier) ([]PostAuthor, error)

	AddComment(ctx context.Context, q database.Querier, c PostComment) error
	FindCommentByID(ctx context.Context, q database.Querier, id uuid.UUID) (PostComment, error)
	// ListPostComments returns every comment on a post, replies included, oldest first.
	ListPostComments(ctx context.Context, q database.Querier, postID uuid.UUID) ([]PostComment, error)
	ListTopLevelComments(ctx context.Context, q database.Querier, postID uuid.UUID, filter FilterQuery) ([]PostComment, error)
	ListCommentReplies(ctx context.Context, q database.Querier, parentID uuid.UUID, filter FilterQuery) ([]PostComment, error)

	// AddEngagement records a view, or a like or dislike. A like or dislike replaces the user's
	// earlier one on the same asset instead of adding to it.
	AddEngagement(ctx context.Context, q database.Querier, e PostEngagement) error
	// RemoveReaction takes back the user's like or dislike on an asset; nothing to remove is not an error.
	RemoveReaction(ctx context.Context, q database.Querier, assetType PostType, assetID, userID uuid.UUID) error
	// ListReactions returns the user's like or dislike on a post and on each of its comments, by asset id.
	ListReactions(ctx context.Context, q database.Querier, userID, postID uuid.UUID) (map[uuid.UUID]ActionType, error)
}
