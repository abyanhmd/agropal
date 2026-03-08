package post

import (
	"time"

	"github.com/google/uuid"
)

type PostModel struct {
	ID           uuid.UUID  `db:"post_id"`
	UserID       uuid.UUID  `db:"user_id"`
	Title        string     `db:"title"`
	Content      string     `db:"content"`
	Slug         string     `db:"slug"`
	CommentCount int        `db:"comment_count"`
	LikeCount    int        `db:"like_count"`
	DislikeCount int        `db:"dislike_count"`
	ViewCount    int        `db:"view_count"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

type CommentModel struct {
	ID           uuid.UUID  `db:"comment_id"`
	PostID       uuid.UUID  `db:"post_id"`
	UserID       uuid.UUID  `db:"user_id"`
	ParentID     *uuid.UUID `db:"parent_id"`
	Content      string     `db:"content"`
	ReplyCount   int        `db:"reply_count"`
	LikeCount    int        `db:"like_count"`
	DislikeCount int        `db:"dislike_count"`
	ViewCount    int        `db:"view_count"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

type EngagementModel struct {
	ID        uuid.UUID `db:"engagement_id"`
	AssetID   uuid.UUID `db:"asset_id"`
	AssetType string    `db:"asset_type"`
	UserID    uuid.UUID `db:"user_id"`
	Type      string    `db:"type"`
	CreatedAt time.Time `db:"created_at"`
}
