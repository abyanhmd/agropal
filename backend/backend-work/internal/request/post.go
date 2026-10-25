package request

import (
	aPost "agropal/internal/app/post"
	dPost "agropal/internal/domain/post"

	"github.com/google/uuid"
)

type PostRequest struct {
	Title   string `json:"title" validate:"required,max=255"`
	Content string `json:"content" validate:"required"`
}

func (r *PostRequest) ToCommand(userID uuid.UUID) aPost.PostCommand {
	return aPost.PostCommand{
		Title:   r.Title,
		Content: r.Content,
		UserID:  userID,
	}
}

type PostCommentRequest struct {
	Content  string     `json:"content" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"` // the comment being answered; omitted for a top-level comment
}

func (r *PostCommentRequest) ToCommand(postID uuid.UUID) aPost.PostCommentCommand {
	return aPost.PostCommentCommand{
		PostID:   postID,
		Content:  r.Content,
		ParentID: r.ParentID,
	}
}

type PostEngagementRequest struct {
	Type string `json:"type"`
}

func (r *PostEngagementRequest) ToCommand(postID uuid.UUID) aPost.PostEngagementCommand {
	return aPost.PostEngagementCommand{
		PostID: postID,
		Type:   dPost.ActionType(r.Type),
	}
}
