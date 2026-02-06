package post

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreatePost(ctx context.Context, post Post) error
	GetOnePost(ctx context.Context, postID uuid.UUID) (PostDetail, error)
	GetManyPosts(ctx context.Context) ([]Post, error)

	CreatePostComment(ctx context.Context, comment PostComment) error
	GetPostComments(ctx context.Context, postID uuid.UUID) ([]PostComment, error)

	CreatePostEngagement(ctx context.Context, engagement PostEngagement) error
}
