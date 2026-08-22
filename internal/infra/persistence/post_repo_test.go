package persistence

import (
	dPost "agropal/internal/domain/post"
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestPostRepository_CreateAndFindByID(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	userID := insertTestUser(t)

	post := dPost.Post{ID: uuid.New(), UserID: userID, Title: "Hello", Content: "World", Slug: uuid.NewString()}
	if err := repo.CreatePost(ctx, testDB, post); err != nil {
		t.Fatalf("CreatePost() error = %v, want nil", err)
	}

	got, err := repo.FindPostByID(ctx, testDB, post.ID)
	if err != nil {
		t.Fatalf("FindPostByID() error = %v, want nil", err)
	}
	if got.ID != post.ID || got.Title != post.Title || got.Content != post.Content {
		t.Errorf("FindPostByID() = %+v, want ID=%v Title=%q Content=%q", got, post.ID, post.Title, post.Content)
	}
	// A fresh post has no engagements or comments yet - counts are
	// computed on read, so they must start at zero rather than being
	// absent/nil.
	if got.LikeCount != 0 || got.DislikeCount != 0 || got.ViewCount != 0 || got.CommentCount != 0 {
		t.Errorf("fresh post counts = %+v, want all zero", got)
	}
}

func TestPostRepository_FindPostBySlug(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	userID := insertTestUser(t)
	slug := uuid.NewString()

	post := dPost.Post{ID: uuid.New(), UserID: userID, Title: "Hello", Content: "World", Slug: slug}
	if err := repo.CreatePost(ctx, testDB, post); err != nil {
		t.Fatalf("CreatePost() error = %v, want nil", err)
	}

	got, err := repo.FindPostBySlug(ctx, testDB, slug)
	if err != nil {
		t.Fatalf("FindPostBySlug() error = %v, want nil", err)
	}
	if got.ID != post.ID {
		t.Errorf("FindPostBySlug().ID = %v, want %v", got.ID, post.ID)
	}
}

func TestPostRepository_EngagementCounts(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	authorID := insertTestUser(t)

	post := dPost.Post{ID: uuid.New(), UserID: authorID, Title: "Hello", Content: "World", Slug: uuid.NewString()}
	if err := repo.CreatePost(ctx, testDB, post); err != nil {
		t.Fatalf("CreatePost() error = %v, want nil", err)
	}

	// Two likes, one dislike, three views from three different users.
	engagements := []struct {
		userID uuid.UUID
		action dPost.ActionType
	}{
		{insertTestUser(t), dPost.ActionLike},
		{insertTestUser(t), dPost.ActionLike},
		{insertTestUser(t), dPost.ActionDislike},
	}
	for _, e := range engagements {
		eng := dPost.PostEngagement{ID: uuid.New(), AssetID: post.ID, AssetType: dPost.TypePost, UserID: e.userID, Type: e.action}
		if err := repo.AddEngagement(ctx, testDB, eng); err != nil {
			t.Fatalf("AddEngagement() error = %v, want nil", err)
		}
	}

	got, err := repo.FindPostByID(ctx, testDB, post.ID)
	if err != nil {
		t.Fatalf("FindPostByID() error = %v, want nil", err)
	}
	if got.LikeCount != 2 {
		t.Errorf("LikeCount = %d, want 2", got.LikeCount)
	}
	if got.DislikeCount != 1 {
		t.Errorf("DislikeCount = %d, want 1", got.DislikeCount)
	}
	if got.ViewCount != 0 {
		t.Errorf("ViewCount = %d, want 0", got.ViewCount)
	}
}

func TestPostRepository_CommentCounts(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	authorID := insertTestUser(t)
	commenterID := insertTestUser(t)

	post := dPost.Post{ID: uuid.New(), UserID: authorID, Title: "Hello", Content: "World", Slug: uuid.NewString()}
	if err := repo.CreatePost(ctx, testDB, post); err != nil {
		t.Fatalf("CreatePost() error = %v, want nil", err)
	}

	topLevel := dPost.PostComment{ID: uuid.New(), PostID: post.ID, UserID: commenterID, Content: "Nice post!"}
	if err := repo.AddComment(ctx, testDB, topLevel); err != nil {
		t.Fatalf("AddComment() error = %v, want nil", err)
	}
	reply := dPost.PostComment{ID: uuid.New(), PostID: post.ID, UserID: commenterID, ParentID: &topLevel.ID, Content: "Agreed!"}
	if err := repo.AddComment(ctx, testDB, reply); err != nil {
		t.Fatalf("AddComment() error = %v, want nil", err)
	}

	// The post's comment_count only counts top-level comments, not replies.
	gotPost, err := repo.FindPostByID(ctx, testDB, post.ID)
	if err != nil {
		t.Fatalf("FindPostByID() error = %v, want nil", err)
	}
	if gotPost.CommentCount != 1 {
		t.Errorf("Post.CommentCount = %d, want 1 (replies excluded)", gotPost.CommentCount)
	}

	topLevelComments, err := repo.ListTopLevelComments(ctx, testDB, post.ID, dPost.FilterQuery{})
	if err != nil {
		t.Fatalf("ListTopLevelComments() error = %v, want nil", err)
	}
	if len(topLevelComments) != 1 {
		t.Fatalf("ListTopLevelComments() returned %d comments, want 1", len(topLevelComments))
	}
	if topLevelComments[0].ReplyCount != 1 {
		t.Errorf("top-level comment ReplyCount = %d, want 1", topLevelComments[0].ReplyCount)
	}

	replies, err := repo.ListCommentReplies(ctx, testDB, topLevel.ID, dPost.FilterQuery{})
	if err != nil {
		t.Fatalf("ListCommentReplies() error = %v, want nil", err)
	}
	if len(replies) != 1 || replies[0].ID != reply.ID {
		t.Fatalf("ListCommentReplies() = %+v, want a single reply with ID %v", replies, reply.ID)
	}
}
