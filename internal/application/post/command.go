package post

import (
	dPost "agropal/internal/domain/post"

	"github.com/google/uuid"
)

type PostCommand struct {
	Title   string
	Content string
	UserID  uuid.UUID
}

type PostCommentCommand struct {
	PostID  uuid.UUID
	Content string
	UserID  uuid.UUID
}

type PostEngagementCommand struct {
	PostID uuid.UUID
	UserID uuid.UUID
	Type   dPost.EngagementType
}

func (c PostCommand) ToDomain(userID uuid.UUID) dPost.Post {
	return dPost.Post{
		Title:   c.Title,
		Content: c.Content,
		UserID:  userID,
	}
}

func (c PostCommentCommand) ToDomain(userID uuid.UUID) dPost.PostComment {
	return dPost.PostComment{
		PostID:  c.PostID,
		Content: c.Content,
		UserID:  userID,
	}
}

func (c PostEngagementCommand) ToDomain(userID uuid.UUID) dPost.PostEngagement {
	return dPost.PostEngagement{
		PostID: c.PostID,
		UserID: userID,
		Type:   c.Type,
	}
}
