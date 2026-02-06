package post

import dPost "agropal/internal/domain/post"

func (m ModelPost) ToDomain() dPost.Post {
	return dPost.Post{
		PostID:  m.PostID,
		Title:   m.Title,
		Content: m.Content,
		UserID:  m.UserID,
		Slug:    m.Slug,
		Like:    m.Like,
		Dislike: m.Dislike,
		View:    m.View,
	}
}

func (m ModelPostComment) ToDomain() dPost.PostComment {
	return dPost.PostComment{
		CommentID: m.CommentID,
		PostID:    m.PostID,
		Content:   m.Content,
		UserID:    m.UserID,
		Like:      m.Like,
		Dislike:   m.Dislike,
		View:      m.View,
	}
}

func (m ModelPostEngagement) ToDomain() dPost.PostEngagement {
	return dPost.PostEngagement{
		EngagementID: m.EngagementID,
		PostID:       m.PostID,
		UserID:       m.UserID,
		Type:         dPost.EngagementType(m.Type),
	}
}

func DomainToModelPost(domain dPost.Post) ModelPost {
	return ModelPost{
		PostID:  domain.PostID,
		Title:   domain.Title,
		Content: domain.Content,
		UserID:  domain.UserID,
		Slug:    domain.Slug,
		Like:    domain.Like,
		Dislike: domain.Dislike,
		View:    domain.View,
	}
}

func DomainToModelComment(domain dPost.PostComment) ModelPostComment {
	return ModelPostComment{
		CommentID: domain.CommentID,
		PostID:    domain.PostID,
		Content:   domain.Content,
		UserID:    domain.UserID,
		Like:      domain.Like,
		Dislike:   domain.Dislike,
		View:      domain.View,
	}
}

func DomainToModelEngagement(domain dPost.PostEngagement) ModelPostEngagement {
	return ModelPostEngagement{
		EngagementID: domain.EngagementID,
		PostID:       domain.PostID,
		UserID:       domain.UserID,
		Type:         string(domain.Type),
	}
}
