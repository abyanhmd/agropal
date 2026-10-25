package post

import (
	"agropal/internal/domain/shared"
	"time"

	"github.com/google/uuid"
)

type PostDetail struct {
	Post
	Comments []PostComment
}

type Post struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Title   string
	Content string
	Slug    string

	LikeCount    int
	DislikeCount int
	ViewCount    int
	CommentCount int

	// AuthorName is the display name of UserID. Other users' profiles are private, so this is the
	// only part of an author the post endpoints expose.
	AuthorName string

	// ViewerReaction is the requester's own like or dislike, empty when they have not reacted.
	// It is filled in when a post is fetched on its own, not on listings.
	ViewerReaction ActionType

	CreatedAt time.Time
	UpdatedAt *time.Time
}

type PostComment struct {
	ID       uuid.UUID
	PostID   uuid.UUID
	UserID   uuid.UUID
	ParentID *uuid.UUID
	Content  string

	ReplyCount   int
	LikeCount    int
	DislikeCount int
	ViewCount    int

	AuthorName string

	ViewerReaction ActionType

	CreatedAt time.Time
	UpdatedAt *time.Time
}

// PostAuthor is someone who has written at least one post.
type PostAuthor struct {
	UserID uuid.UUID
	Name   string
}

type PostEngagement struct {
	ID        uuid.UUID
	AssetID   uuid.UUID
	AssetType PostType
	UserID    uuid.UUID
	Type      ActionType
	CreatedAt time.Time
}

type PostType string

type ActionType string

const (
	TypePost    PostType = "POST"
	TypeComment PostType = "COMMENT"

	ActionLike    ActionType = "LIKE"
	ActionDislike ActionType = "DISLIKE"
	ActionView    ActionType = "VIEW"
)

type FilterQuery struct {
	UserIDs    []uuid.UUID // any of these authors; empty means everyone
	Slug       *string
	Search     *string // matched against title and content
	Pagination *shared.PaginationQuery
	Sorting    *shared.SortQuery
}

// Sort columns a client may ask for. The sort comes from the query string and ends up in an
// ORDER BY, so anything outside these lists must be rejected, never passed on.
var (
	PostSortColumns    = map[string]struct{}{"created_at": {}, "updated_at": {}, "title": {}, "like_count": {}, "dislike_count": {}, "view_count": {}, "comment_count": {}}
	CommentSortColumns = map[string]struct{}{"created_at": {}, "updated_at": {}, "like_count": {}, "dislike_count": {}, "view_count": {}, "reply_count": {}}
)

// ValidSort reports whether a sort request names an allowed column and a direction. A nil sort is valid.
func ValidSort(sort *shared.SortQuery, allowed map[string]struct{}) bool {
	if sort == nil {
		return true
	}
	_, ok := allowed[sort.Column]
	return ok && (sort.Dir == "asc" || sort.Dir == "desc")
}

var (
	SuccessCreated shared.SuccessKey = "POST.CREATE_SUCCESS"
	SuccessUpdated shared.SuccessKey = "POST.UPDATE_SUCCESS"
	SuccessDeleted shared.SuccessKey = "POST.DELETE_SUCCESS"

	SuccessCommentCreated shared.SuccessKey = "POST.COMMENT_CREATE_SUCCESS"
	SuccessCommentUpdated shared.SuccessKey = "POST.COMMENT_UPDATE_SUCCESS"
	SuccessCommentDeleted shared.SuccessKey = "POST.COMMENT_DELETE_SUCCESS"

	SuccessEngagementCreated shared.SuccessKey = "POST.ENGAGEMENT_CREATE_SUCCESS"
	SuccessEngagementDeleted shared.SuccessKey = "POST.ENGAGEMENT_DELETE_SUCCESS"
)
