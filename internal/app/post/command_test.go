package post

import (
	dPost "agropal/internal/domain/post"
	"testing"

	"github.com/google/uuid"
)

func TestPostCommand_ToDomain(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	cmd := PostCommand{Title: "Hello", Content: "World", UserID: userID}

	got := cmd.ToDomain(postID, userID)

	want := dPost.Post{
		ID:      postID,
		UserID:  userID,
		Title:   "Hello",
		Content: "World",
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}

func TestPostCommentCommand_ToDomain(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	cmd := PostCommentCommand{PostID: postID, Content: "Nice post!"}

	got := cmd.ToDomain(userID)

	want := dPost.PostComment{
		PostID:  postID,
		Content: "Nice post!",
		UserID:  userID,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}

func TestPostEngagementCommand_ToDomain(t *testing.T) {
	postID := uuid.New()
	userID := uuid.New()

	cmd := PostEngagementCommand{PostID: postID, Type: dPost.ActionLike}

	got := cmd.ToDomain(userID)

	// PostEngagementCommand.ToDomain maps PostID into the domain
	// PostEngagement's ID field (not AssetID) - this test locks in that
	// existing mapping rather than the more intuitive one.
	want := dPost.PostEngagement{
		ID:     postID,
		UserID: userID,
		Type:   dPost.ActionLike,
	}

	if got != want {
		t.Errorf("ToDomain() = %+v, want %+v", got, want)
	}
}
