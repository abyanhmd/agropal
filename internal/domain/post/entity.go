package post

import "github.com/google/uuid"

type PostDetail struct {
	Post
	Comments []PostComment
}

type Post struct {
	PostID  uuid.UUID
	Title   string
	Content string
	UserID  uuid.UUID
	Slug    string
	Like    int
	Dislike int
	View    int
}

type PostComment struct {
	CommentID uuid.UUID
	PostID    uuid.UUID
	Content   string
	UserID    uuid.UUID
	Like      int
	Dislike   int
	View      int
}

type PostEngagement struct {
	EngagementID uuid.UUID
	PostID       uuid.UUID
	UserID       uuid.UUID
	Type         EngagementType
}

type EngagementType string

const (
	EngagementPostLike    EngagementType = "post_like"
	EngagementPostDislike EngagementType = "post_dislike"
	EngagementPostView    EngagementType = "post_view"

	EngagementPostCommentLike    EngagementType = "post_comment_like"
	EngagementPostCommentDislike EngagementType = "post_comment_dislike"
	EngagementPostCommentView    EngagementType = "post_comment_view"
)
