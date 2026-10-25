package response

import (
	"time"

	"github.com/google/uuid"
)

type PostResponse struct {
	PostID       uuid.UUID `json:"post_id"`
	UserID       uuid.UUID `json:"user_id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	Slug         string    `json:"slug"`
	LikeCount    int       `json:"like_count"`
	DislikeCount int       `json:"dislike_count"`
	ViewCount    int       `json:"view_count"`
	CommentCount int       `json:"comment_count"`
	AuthorName   string    `json:"author_name"`
	// ViewerReaction is LIKE or DISLIKE when the requester reacted; only sent when a post is fetched on its own
	ViewerReaction string     `json:"viewer_reaction,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

type PostCommentResponse struct {
	CommentID      uuid.UUID  `json:"comment_id"`
	PostID         uuid.UUID  `json:"post_id"`
	UserID         uuid.UUID  `json:"user_id"`
	ParentID       *uuid.UUID `json:"parent_id,omitempty"`
	Content        string     `json:"content"`
	ReplyCount     int        `json:"reply_count"`
	LikeCount      int        `json:"like_count"`
	DislikeCount   int        `json:"dislike_count"`
	ViewCount      int        `json:"view_count"`
	AuthorName     string     `json:"author_name"`
	ViewerReaction string     `json:"viewer_reaction,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

type PostAuthorResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

type PostDetailResponse struct {
	Post     PostResponse          `json:"post"`
	Comments []PostCommentResponse `json:"comments"`
}
