package persistence

import (
	dPost "agropal/internal/domain/post"
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
	AuthorName   string     `db:"author_name"`
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
	AuthorName   string     `db:"author_name"`
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

func ToPostModel(domain dPost.Post) PostModel {
	return PostModel{
		ID:           domain.ID,
		UserID:       domain.UserID,
		Title:        domain.Title,
		Content:      domain.Content,
		Slug:         domain.Slug,
		CommentCount: domain.CommentCount,
		LikeCount:    domain.LikeCount,
		DislikeCount: domain.DislikeCount,
		ViewCount:    domain.ViewCount,
		CreatedAt:    domain.CreatedAt,
		UpdatedAt:    domain.UpdatedAt,
	}
}

func ToCommentModel(domain dPost.PostComment) CommentModel {
	return CommentModel{
		ID:           domain.ID,
		PostID:       domain.PostID,
		UserID:       domain.UserID,
		ParentID:     domain.ParentID,
		Content:      domain.Content,
		ReplyCount:   domain.ReplyCount,
		LikeCount:    domain.LikeCount,
		DislikeCount: domain.DislikeCount,
		ViewCount:    domain.ViewCount,
		CreatedAt:    domain.CreatedAt,
		UpdatedAt:    domain.UpdatedAt,
	}
}

func ToEngagementModel(domain dPost.PostEngagement) EngagementModel {
	return EngagementModel{
		ID:        domain.ID,
		AssetID:   domain.AssetID,
		AssetType: string(domain.AssetType),
		UserID:    domain.UserID,
		Type:      string(domain.Type),
		CreatedAt: domain.CreatedAt,
	}
}

func (m PostModel) ToDomain() dPost.Post {
	return dPost.Post{
		ID:           m.ID,
		UserID:       m.UserID,
		Title:        m.Title,
		Content:      m.Content,
		Slug:         m.Slug,
		CommentCount: m.CommentCount,
		AuthorName:   m.AuthorName,
		LikeCount:    m.LikeCount,
		DislikeCount: m.DislikeCount,
		ViewCount:    m.ViewCount,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (m CommentModel) ToDomain() dPost.PostComment {
	return dPost.PostComment{
		ID:           m.ID,
		PostID:       m.PostID,
		UserID:       m.UserID,
		ParentID:     m.ParentID,
		Content:      m.Content,
		ReplyCount:   m.ReplyCount,
		LikeCount:    m.LikeCount,
		DislikeCount: m.DislikeCount,
		ViewCount:    m.ViewCount,
		AuthorName:   m.AuthorName,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func (m EngagementModel) ToDomain() dPost.PostEngagement {
	return dPost.PostEngagement{
		ID:        m.ID,
		AssetID:   m.AssetID,
		AssetType: dPost.PostType(m.AssetType),
		UserID:    m.UserID,
		Type:      dPost.ActionType(m.Type),
		CreatedAt: m.CreatedAt,
	}
}
