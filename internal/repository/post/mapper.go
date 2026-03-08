package post

import dPost "agropal/internal/domain/post"

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
