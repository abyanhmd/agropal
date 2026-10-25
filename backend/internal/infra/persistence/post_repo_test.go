package persistence

import (
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

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

func TestPostRepository_ListSearchSortCountAndAuthors(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	authorA := insertTestUser(t)
	authorB := insertTestUser(t)

	// A word no other test's data contains, so the search only matches these posts
	marker := "zq" + strings.ReplaceAll(uuid.NewString(), "-", "")
	posts := []dPost.Post{
		{ID: uuid.New(), UserID: authorA, Title: marker + " cover crops", Content: "clay", Slug: uuid.NewString()},
		{ID: uuid.New(), UserID: authorA, Title: "Drip", Content: "irrigation " + marker, Slug: uuid.NewString()},
		{ID: uuid.New(), UserID: authorB, Title: marker + " aphids", Content: "no spray", Slug: uuid.NewString()},
	}
	for _, p := range posts {
		if err := repo.CreatePost(ctx, testDB, p); err != nil {
			t.Fatalf("CreatePost() error = %v, want nil", err)
		}
	}
	// One like on the second post, so sorting by likes puts it first
	if err := repo.AddEngagement(ctx, testDB, dPost.PostEngagement{ID: uuid.New(), AssetID: posts[1].ID, AssetType: dPost.TypePost, UserID: authorB, Type: dPost.ActionLike}); err != nil {
		t.Fatalf("AddEngagement() error = %v, want nil", err)
	}

	search := marker
	filter := dPost.FilterQuery{Search: &search, Sorting: &shared.SortQuery{Column: "like_count", Dir: "desc"}}

	got, err := repo.ListPosts(ctx, testDB, filter)
	if err != nil {
		t.Fatalf("ListPosts() error = %v, want nil", err)
	}
	if len(got) != 3 {
		t.Fatalf("ListPosts() returned %d posts, want 3 (title and content both match)", len(got))
	}
	if got[0].ID != posts[1].ID || got[0].LikeCount != 1 {
		t.Errorf("most liked first: got %v with %d likes, want %v with 1", got[0].ID, got[0].LikeCount, posts[1].ID)
	}
	if got[0].AuthorName != "Test User" {
		t.Errorf("AuthorName = %q, want %q", got[0].AuthorName, "Test User")
	}

	// The total ignores paging, and the page slices the same ordering
	filter.Pagination = &shared.PaginationQuery{Limit: 2, Offset: 2}
	page, err := repo.ListPosts(ctx, testDB, filter)
	if err != nil {
		t.Fatalf("ListPosts() paged error = %v, want nil", err)
	}
	if len(page) != 1 {
		t.Errorf("last page has %d posts, want 1", len(page))
	}
	total, err := repo.CountPosts(ctx, testDB, filter)
	if err != nil {
		t.Fatalf("CountPosts() error = %v, want nil", err)
	}
	if total != 3 {
		t.Errorf("CountPosts() = %d, want 3", total)
	}

	// Several authors at once
	filter.Pagination = nil
	filter.UserIDs = []uuid.UUID{authorB}
	byB, err := repo.ListPosts(ctx, testDB, filter)
	if err != nil {
		t.Fatalf("ListPosts() by author error = %v, want nil", err)
	}
	if len(byB) != 1 || byB[0].UserID != authorB {
		t.Errorf("ListPosts() by author = %+v, want only author B's post", byB)
	}

	// Wildcards in the search are literal
	percent := "%"
	none, err := repo.CountPosts(ctx, testDB, dPost.FilterQuery{Search: &percent, UserIDs: []uuid.UUID{authorA, authorB}})
	if err != nil {
		t.Fatalf("CountPosts() wildcard error = %v, want nil", err)
	}
	if none != 0 {
		t.Errorf("a search for %q matched %d posts, want 0", percent, none)
	}

	// An unknown sort column falls back to the default instead of reaching the query
	bad := dPost.FilterQuery{UserIDs: []uuid.UUID{authorA}, Sorting: &shared.SortQuery{Column: "1; DROP TABLE posts", Dir: "asc"}}
	if _, err := repo.ListPosts(ctx, testDB, bad); err != nil {
		t.Errorf("ListPosts() with an unknown sort column error = %v, want nil (default order)", err)
	}

	authors, err := repo.ListPostAuthors(ctx, testDB)
	if err != nil {
		t.Fatalf("ListPostAuthors() error = %v, want nil", err)
	}
	seen := map[uuid.UUID]bool{}
	for _, a := range authors {
		seen[a.UserID] = true
	}
	if !seen[authorA] || !seen[authorB] {
		t.Errorf("ListPostAuthors() = %+v, want both authors", authors)
	}
}

func TestPostRepository_ListPostCommentsAndFindComment(t *testing.T) {
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
	now := time.Now().UTC()
	topLevel := dPost.PostComment{ID: uuid.New(), PostID: post.ID, UserID: authorID, Content: "First", CreatedAt: now}
	reply := dPost.PostComment{ID: uuid.New(), PostID: post.ID, UserID: authorID, ParentID: &topLevel.ID, Content: "Second", CreatedAt: now.Add(time.Second)}
	for _, c := range []dPost.PostComment{topLevel, reply} {
		if err := repo.AddComment(ctx, testDB, c); err != nil {
			t.Fatalf("AddComment() error = %v, want nil", err)
		}
	}

	all, err := repo.ListPostComments(ctx, testDB, post.ID)
	if err != nil {
		t.Fatalf("ListPostComments() error = %v, want nil", err)
	}
	if len(all) != 2 || all[0].ID != topLevel.ID || all[1].ID != reply.ID {
		t.Fatalf("ListPostComments() = %+v, want the comment then its reply, oldest first", all)
	}
	if all[1].ParentID == nil || *all[1].ParentID != topLevel.ID || all[0].AuthorName != "Test User" {
		t.Errorf("reply parent / author name wrong: %+v", all)
	}

	found, err := repo.FindCommentByID(ctx, testDB, reply.ID)
	if err != nil || found.PostID != post.ID {
		t.Errorf("FindCommentByID() = %+v, %v, want the reply on post %v", found, err, post.ID)
	}
	if _, err := repo.FindCommentByID(ctx, testDB, uuid.New()); !errors.Is(err, dPost.ErrCommentNotFound) {
		t.Errorf("FindCommentByID() unknown id error = %v, want %v", err, dPost.ErrCommentNotFound)
	}
}

func TestPostRepository_OneReactionPerUser(t *testing.T) {
	if testDB == nil {
		t.Skip("no test database available (see main_test.go)")
	}
	ctx := context.Background()
	repo := NewPostRepository(testLogger)
	authorID := insertTestUser(t)
	otherID := insertTestUser(t)

	post := dPost.Post{ID: uuid.New(), UserID: authorID, Title: "Hello", Content: "World", Slug: uuid.NewString()}
	if err := repo.CreatePost(ctx, testDB, post); err != nil {
		t.Fatalf("CreatePost() error = %v, want nil", err)
	}
	comment := dPost.PostComment{ID: uuid.New(), PostID: post.ID, UserID: authorID, Content: "Nice"}
	if err := repo.AddComment(ctx, testDB, comment); err != nil {
		t.Fatalf("AddComment() error = %v, want nil", err)
	}

	react := func(assetType dPost.PostType, assetID, userID uuid.UUID, action dPost.ActionType) {
		t.Helper()
		e := dPost.PostEngagement{ID: uuid.New(), AssetID: assetID, AssetType: assetType, UserID: userID, Type: action}
		if err := repo.AddEngagement(ctx, testDB, e); err != nil {
			t.Fatalf("AddEngagement(%s) error = %v, want nil", action, err)
		}
	}
	counts := func() dPost.Post {
		t.Helper()
		got, err := repo.FindPostByID(ctx, testDB, post.ID)
		if err != nil {
			t.Fatalf("FindPostByID() error = %v, want nil", err)
		}
		return got
	}

	// Liking twice is still one like
	react(dPost.TypePost, post.ID, authorID, dPost.ActionLike)
	react(dPost.TypePost, post.ID, authorID, dPost.ActionLike)
	react(dPost.TypePost, post.ID, otherID, dPost.ActionLike)
	if got := counts(); got.LikeCount != 2 || got.DislikeCount != 0 {
		t.Fatalf("after repeated likes: likes=%d dislikes=%d, want 2 and 0", got.LikeCount, got.DislikeCount)
	}

	// Switching moves the reaction instead of adding one
	react(dPost.TypePost, post.ID, authorID, dPost.ActionDislike)
	if got := counts(); got.LikeCount != 1 || got.DislikeCount != 1 {
		t.Fatalf("after switching: likes=%d dislikes=%d, want 1 and 1", got.LikeCount, got.DislikeCount)
	}

	// Views are not reactions: every one counts
	react(dPost.TypePost, post.ID, authorID, dPost.ActionView)
	react(dPost.TypePost, post.ID, authorID, dPost.ActionView)
	if got := counts(); got.ViewCount != 2 {
		t.Errorf("ViewCount = %d, want 2", got.ViewCount)
	}

	react(dPost.TypeComment, comment.ID, authorID, dPost.ActionLike)

	reactions, err := repo.ListReactions(ctx, testDB, authorID, post.ID)
	if err != nil {
		t.Fatalf("ListReactions() error = %v, want nil", err)
	}
	if reactions[post.ID] != dPost.ActionDislike || reactions[comment.ID] != dPost.ActionLike || len(reactions) != 2 {
		t.Errorf("ListReactions() = %v, want a dislike on the post and a like on the comment", reactions)
	}

	// Taking it back, twice, is fine
	for i := 0; i < 2; i++ {
		if err := repo.RemoveReaction(ctx, testDB, dPost.TypePost, post.ID, authorID); err != nil {
			t.Fatalf("RemoveReaction() error = %v, want nil", err)
		}
	}
	if got := counts(); got.LikeCount != 1 || got.DislikeCount != 0 || got.ViewCount != 2 {
		t.Errorf("after removing: likes=%d dislikes=%d views=%d, want 1, 0 and 2 (the other user's like and the views stay)", got.LikeCount, got.DislikeCount, got.ViewCount)
	}
}
