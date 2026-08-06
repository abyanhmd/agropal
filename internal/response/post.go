package response

import (
	"time"

	"github.com/google/uuid"
)

type PostResponse struct {
	PostID       uuid.UUID  `json:"post_id"`
	UserID       uuid.UUID  `json:"user_id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Slug         string     `json:"slug"`
	LikeCount    int        `json:"like_count"`
	DislikeCount int        `json:"dislike_count"`
	ViewCount    int        `json:"view_count"`
	CommentCount int        `json:"comment_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type PostCommentResponse struct {
	CommentID    uuid.UUID  `json:"comment_id"`
	PostID       uuid.UUID  `json:"post_id"`
	UserID       uuid.UUID  `json:"user_id"`
	ParentID     *uuid.UUID `json:"parent_id,omitempty"`
	Content      string     `json:"content"`
	ReplyCount   int        `json:"reply_count"`
	LikeCount    int        `json:"like_count"`
	DislikeCount int        `json:"dislike_count"`
	ViewCount    int        `json:"view_count"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

type PostDetailResponse struct {
	Post     PostResponse          `json:"post"`
	Comments []PostCommentResponse `json:"comments"`
}
