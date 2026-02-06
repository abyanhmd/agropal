package post

import "github.com/google/uuid"

type ModelPost struct {
	PostID  uuid.UUID `db:"post_id"`
	Title   string    `db:"title"`
	Content string    `db:"content"`
	UserID  uuid.UUID `db:"user_id"`
	Slug    string    `db:"slug"`
	Like    int       `db:"like"`
	Dislike int       `db:"dislike"`
	View    int       `db:"view"`
}

type ModelPostComment struct {
	CommentID uuid.UUID `db:"comment_id"`
	PostID    uuid.UUID `db:"post_id"`
	Content   string    `db:"content"`
	UserID    uuid.UUID `db:"user_id"`
	Like      int       `db:"like"`
	Dislike   int       `db:"dislike"`
	View      int       `db:"view"`
}

type ModelPostEngagement struct {
	EngagementID uuid.UUID `db:"engagement_id"`
	PostID       uuid.UUID `db:"post_id"`
	UserID       uuid.UUID `db:"user_id"`
	Type         string    `db:"type"`
}
