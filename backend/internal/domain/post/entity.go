package post

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type PostDetail struct {
	Post
	Comments []PostComment
}

type Post struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Title   string
	Content string
	Slug    string

	LikeCount    int
	DislikeCount int
	ViewCount    int
	CommentCount int

	CreatedAt time.Time
	UpdatedAt *time.Time
}

type PostComment struct {
	ID       uuid.UUID
	PostID   uuid.UUID
	UserID   uuid.UUID
	ParentID *uuid.UUID
	Content  string

	ReplyCount   int
	LikeCount    int
	DislikeCount int
	ViewCount    int

	CreatedAt time.Time
	UpdatedAt *time.Time
}

type PostEngagement struct {
	ID        uuid.UUID
	AssetID   uuid.UUID
	AssetType PostType
	UserID    uuid.UUID
	Type      ActionType
	CreatedAt time.Time
}

type PostType string

type ActionType string

const (
	TypePost    PostType = "POST"
	TypeComment PostType = "COMMENT"

	ActionLike    ActionType = "LIKE"
	ActionDislike ActionType = "DISLIKE"
	ActionView    ActionType = "VIEW"
)

type FilterQuery struct {
	UserID     *uuid.UUID
	Slug       *string
	Pagination *shared.PaginationQuery
	Sorting    *shared.SortQuery
}

var (
	SuccessCreated shared.SuccessKey = "POST.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "POST.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "POST.DELETE_SUCCESS"

	SuccessCommentCreated shared.SuccessKey = "POST.COMMENT_CREATE_SUCCESS"
	SuccessCommentUpdated shared.SuccessKey = "POST.COMMENT_UPDATE_SUCCESS"
	SuccessCommentDeleted shared.SuccessKey = "POST.COMMENT_DELETE_SUCCESS"
)
