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
	PostID   uuid.UUID
	Content  string
	UserID   uuid.UUID
	ParentID *uuid.UUID
}

type PostEngagementCommand struct {
	PostID uuid.UUID
	UserID uuid.UUID
	Type   dPost.ActionType
}

func (c PostCommand) ToDomain(postID, userID uuid.UUID) dPost.Post {
	return dPost.Post{
		ID:      postID,
		UserID:  userID,
		Title:   c.Title,
		Content: c.Content,
	}
}

func (c PostCommentCommand) ToDomain(userID uuid.UUID) dPost.PostComment {
	return dPost.PostComment{
		PostID:   c.PostID,
		Content:  c.Content,
		UserID:   userID,
		ParentID: c.ParentID,
	}
}

func (c PostEngagementCommand) ToDomain(userID uuid.UUID) dPost.PostEngagement {
	return dPost.PostEngagement{
		ID:     c.PostID,
		UserID: userID,
		Type:   c.Type,
	}
}
