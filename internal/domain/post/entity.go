package post

import (
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

	CommentCount int
	LikeCount    int
	DislikeCount int
	ViewCount    int

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
