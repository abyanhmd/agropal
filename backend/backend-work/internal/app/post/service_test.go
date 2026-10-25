package post

import (
	"agropal/config"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"errors"
	"testing"
	"strings"
	"time"

	"github.com/google/uuid"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier
// instead of opening a real transaction. Safe here because the fake
// repositories below never touch the Querier they're given - they stand in
// for the whole repository, not just the SQL layer underneath it.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier {
	return nil
}

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by post.Service.
type fakeUserRepository struct {
	findUserByIDFunc func(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error)
}

func (f *fakeUserRepository) CreateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: CreateUser not stubbed")
}

func (f *fakeUserRepository) FindUserByID(ctx context.Context, q database.Querier, userID uuid.UUID) (dUser.User, error) {
	if f.findUserByIDFunc == nil {
		panic("fakeUserRepository: findUserByIDFunc not stubbed")
	}
	return f.findUserByIDFunc(ctx, q, userID)
}

func (f *fakeUserRepository) FindUserByEmail(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByEmail not stubbed")
}

func (f *fakeUserRepository) FindUserByUsername(context.Context, database.Querier, string) (dUser.User, error) {
	panic("fakeUserRepository: FindUserByUsername not stubbed")
}

func (f *fakeUserRepository) ListUsers(context.Context, database.Querier) ([]dUser.User, error) {
	panic("fakeUserRepository: ListUsers not stubbed")
}

func (f *fakeUserRepository) UpdateUser(context.Context, database.Querier, dUser.User) error {
	panic("fakeUserRepository: UpdateUser not stubbed")
}

func (f *fakeUserRepository) UpdateUserPassword(context.Context, database.Querier, uuid.UUID, string) error {
	panic("fakeUserRepository: UpdateUserPassword not stubbed")
}

func (f *fakeUserRepository) VerifyUserEmail(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: VerifyUserEmail not stubbed")
}

func (f *fakeUserRepository) SoftDeleteUser(context.Context, database.Querier, uuid.UUID, time.Time) error {
	panic("fakeUserRepository: SoftDeleteUser not stubbed")
}

func (f *fakeUserRepository) DeleteUser(context.Context, database.Querier, uuid.UUID) error {
	panic("fakeUserRepository: DeleteUser not stubbed")
}

// fakePostRepository implements dPost.Repository. Only the methods
// post.Service actually calls are wired to a *Func field; everything else
// panics if called.
type fakePostRepository struct {
	createPostFunc         func(ctx context.Context, q database.Querier, p dPost.Post) error
	findPostByIDFunc       func(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error)
	listPostsFunc          func(ctx context.Context, q database.Querier, filter dPost.FilterQuery) ([]dPost.Post, error)
	addCommentFunc         func(ctx context.Context, q database.Querier, c dPost.PostComment) error
	countPostsFunc         func(ctx context.Context, q database.Querier, filter dPost.FilterQuery) (int, error)
	findCommentByIDFunc    func(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.PostComment, error)
	listPostCommentsFunc   func(ctx context.Context, q database.Querier, postID uuid.UUID) ([]dPost.PostComment, error)
	addEngagementFunc      func(ctx context.Context, q database.Querier, e dPost.PostEngagement) error
	removeReactionFunc     func(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID, userID uuid.UUID) error
	listReactionsFunc      func(ctx context.Context, q database.Querier, userID, postID uuid.UUID) (map[uuid.UUID]dPost.ActionType, error)
}

func (f *fakePostRepository) CreatePost(ctx context.Context, q database.Querier, p dPost.Post) error {
	if f.createPostFunc == nil {
		panic("fakePostRepository: createPostFunc not stubbed")
	}
	return f.createPostFunc(ctx, q, p)
}

func (f *fakePostRepository) UpdatePost(context.Context, database.Querier, dPost.Post) error {
	panic("fakePostRepository: UpdatePost not stubbed")
}

func (f *fakePostRepository) DeletePost(context.Context, database.Querier, uuid.UUID) error {
	panic("fakePostRepository: DeletePost not stubbed")
}

func (f *fakePostRepository) FindPostByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error) {
	if f.findPostByIDFunc == nil {
		panic("fakePostRepository: findPostByIDFunc not stubbed")
	}
	return f.findPostByIDFunc(ctx, q, id)
}

func (f *fakePostRepository) FindPostBySlug(context.Context, database.Querier, string) (dPost.Post, error) {
	panic("fakePostRepository: FindPostBySlug not stubbed")
}

func (f *fakePostRepository) ListPosts(ctx context.Context, q database.Querier, filter dPost.FilterQuery) ([]dPost.Post, error) {
	if f.listPostsFunc == nil {
		panic("fakePostRepository: listPostsFunc not stubbed")
	}
	return f.listPostsFunc(ctx, q, filter)
}

func (f *fakePostRepository) AddComment(ctx context.Context, q database.Querier, c dPost.PostComment) error {
	if f.addCommentFunc == nil {
		panic("fakePostRepository: addCommentFunc not stubbed")
	}
	return f.addCommentFunc(ctx, q, c)
}

func (f *fakePostRepository) ListTopLevelComments(context.Context, database.Querier, uuid.UUID, dPost.FilterQuery) ([]dPost.PostComment, error) {
	panic("fakePostRepository: ListTopLevelComments not stubbed")
}

func (f *fakePostRepository) ListCommentReplies(context.Context, database.Querier, uuid.UUID, dPost.FilterQuery) ([]dPost.PostComment, error) {
	panic("fakePostRepository: ListCommentReplies not stubbed")
}

func (f *fakePostRepository) CountPosts(ctx context.Context, q database.Querier, filter dPost.FilterQuery) (int, error) {
	if f.countPostsFunc == nil {
		panic("fakePostRepository: countPostsFunc not stubbed")
	}
	return f.countPostsFunc(ctx, q, filter)
}

func (f *fakePostRepository) ListPostAuthors(context.Context, database.Querier) ([]dPost.PostAuthor, error) {
	panic("fakePostRepository: ListPostAuthors not stubbed")
}

func (f *fakePostRepository) FindCommentByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.PostComment, error) {
	if f.findCommentByIDFunc == nil {
		panic("fakePostRepository: findCommentByIDFunc not stubbed")
	}
	return f.findCommentByIDFunc(ctx, q, id)
}

func (f *fakePostRepository) ListPostComments(ctx context.Context, q database.Querier, postID uuid.UUID) ([]dPost.PostComment, error) {
	if f.listPostCommentsFunc == nil {
		panic("fakePostRepository: listPostCommentsFunc not stubbed")
	}
	return f.listPostCommentsFunc(ctx, q, postID)
}

func (f *fakePostRepository) RemoveReaction(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID, userID uuid.UUID) error {
	if f.removeReactionFunc == nil {
		panic("fakePostRepository: removeReactionFunc not stubbed")
	}
	return f.removeReactionFunc(ctx, q, assetType, assetID, userID)
}

func (f *fakePostRepository) ListReactions(ctx context.Context, q database.Querier, userID, postID uuid.UUID) (map[uuid.UUID]dPost.ActionType, error) {
	if f.listReactionsFunc == nil {
		panic("fakePostRepository: listReactionsFunc not stubbed")
	}
	return f.listReactionsFunc(ctx, q, userID, postID)
}

func (f *fakePostRepository) AddEngagement(ctx context.Context, q database.Querier, e dPost.PostEngagement) error {
	if f.addEngagementFunc == nil {
		panic("fakePostRepository: addEngagementFunc not stubbed")
	}
	return f.addEngagementFunc(ctx, q, e)
}

// fakeNoteRepository implements dNote.Repository. Only CreateNote is
// exercised by post.Service.
type fakeNoteRepository struct {
	createNoteFunc func(ctx context.Context, q database.Querier, note dNote.Note) error
}

func (f *fakeNoteRepository) CreateNote(ctx context.Context, q database.Querier, note dNote.Note) error {
	if f.createNoteFunc == nil {
		panic("fakeNoteRepository: createNoteFunc not stubbed")
	}
	return f.createNoteFunc(ctx, q, note)
}

func (f *fakeNoteRepository) CreateManyNotes(context.Context, database.Querier, []dNote.Note) error {
	panic("fakeNoteRepository: CreateManyNotes not stubbed")
}

func (f *fakeNoteRepository) ListNotes(context.Context, database.Querier, dNote.NoteType, []uuid.UUID, map[string]string) ([]dNote.Note, error) {
	panic("fakeNoteRepository: ListNotes not stubbed")
}

func (f *fakeNoteRepository) DeleteNotes(context.Context, database.Querier, []uuid.UUID) error {
	panic("fakeNoteRepository: DeleteNotes not stubbed")
}

/* ==================== Helpers ====================== */

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func ctxWithUserAndRole(userID uuid.UUID, role config.Role) context.Context {
	ctx := ctxWithUser(userID)
	return shared.WithMetadata(ctx, shared.KeyRole, string(role))
}

/* ==================== Tests: CreatePost ==================== */

func TestService_CreatePost(t *testing.T) {
	selfID := uuid.New()
	otherID := uuid.New()
	adminID := uuid.New()

	t.Run("creates a post for self and records a create note", func(t *testing.T) {
		var createdPost dPost.Post
		var note dNote.Note
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: selfID, Account: dUser.AccountInfo{Name: "Alice"}}, nil
			},
		}
		postRepo := &fakePostRepository{
			createPostFunc: func(_ context.Context, _ database.Querier, p dPost.Post) error {
				createdPost = p
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(_ context.Context, _ database.Querier, n dNote.Note) error {
				note = n
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, postRepo, noteRepo)

		cmd := PostCommand{Title: "Hello, World!", Content: "World", UserID: selfID}
		returned, err := svc.CreatePost(ctxWithUser(selfID), cmd)
		if err != nil {
			t.Fatalf("CreatePost() error = %v, want nil", err)
		}
		if returned.ID != createdPost.ID || returned.Slug != createdPost.Slug {
			t.Errorf("CreatePost() = %v %q, want the stored post %v %q", returned.ID, returned.Slug, createdPost.ID, createdPost.Slug)
		}
		if !strings.HasPrefix(createdPost.Slug, "hello-world-") {
			t.Errorf("Slug = %q, want it built from the title", createdPost.Slug)
		}
		if time.Since(createdPost.CreatedAt) > time.Minute || createdPost.CreatedAt.IsZero() {
			t.Errorf("CreatedAt = %v, want the time of creation", createdPost.CreatedAt)
		}
		if createdPost.UserID != selfID || createdPost.Title != "Hello, World!" {
			t.Errorf("createdPost = %+v, want UserID=%v Title=\"Hello, World!\"", createdPost, selfID)
		}
		if createdPost.ID == uuid.Nil {
			t.Error("createdPost.ID is zero, want a generated UUID")
		}
		if note.AssetID != createdPost.ID || note.AuthorID != selfID || note.AuthorName != "Alice" {
			t.Errorf("note = %+v, want AssetID=%v AuthorID=%v AuthorName=Alice", note, createdPost.ID, selfID)
		}
	})

	t.Run("forbids creating a post on behalf of another non-admin user", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherID}, nil
			},
		}
		postRepo := &fakePostRepository{
			createPostFunc: func(context.Context, database.Querier, dPost.Post) error {
				t.Fatal("CreatePost should not be called when the requester is forbidden")
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, postRepo, &fakeNoteRepository{})

		cmd := PostCommand{Title: "Hello", Content: "World", UserID: otherID}
		_, err := svc.CreatePost(ctxWithUser(selfID), cmd)
		if !errors.Is(err, shared.ErrForbidden) {
			t.Errorf("CreatePost() error = %v, want %v", err, shared.ErrForbidden)
		}
	})

	t.Run("admin can create a post on behalf of another user", func(t *testing.T) {
		var createdPost dPost.Post
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: otherID, Account: dUser.AccountInfo{Name: "Bob"}}, nil
			},
		}
		postRepo := &fakePostRepository{
			createPostFunc: func(_ context.Context, _ database.Querier, p dPost.Post) error {
				createdPost = p
				return nil
			},
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(context.Context, database.Querier, dNote.Note) error {
				return nil
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, postRepo, noteRepo)

		cmd := PostCommand{Title: "Hello", Content: "World", UserID: otherID}
		_, err := svc.CreatePost(ctxWithUserAndRole(adminID, config.RoleAdmin), cmd)
		if err != nil {
			t.Fatalf("CreatePost() error = %v, want nil", err)
		}
		if createdPost.UserID != otherID {
			t.Errorf("createdPost.UserID = %v, want %v", createdPost.UserID, otherID)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		wantErr := errors.New("insert failed")
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: selfID}, nil
			},
		}
		postRepo := &fakePostRepository{
			createPostFunc: func(context.Context, database.Querier, dPost.Post) error {
				return wantErr
			},
		}
		svc := NewService(fakeTransactor{}, userRepo, postRepo, &fakeNoteRepository{})

		cmd := PostCommand{Title: "Hello", Content: "World", UserID: selfID}
		_, err := svc.CreatePost(ctxWithUser(selfID), cmd)
		if !errors.Is(err, wantErr) {
			t.Errorf("CreatePost() error = %v, want %v", err, wantErr)
		}
	})
}

/* ==================== Tests: GetOnePost / GetManyPosts ==================== */

func TestService_GetOnePost(t *testing.T) {
	postID := uuid.New()
	want := dPost.Post{ID: postID, Title: "Hello"}
	wantComments := []dPost.PostComment{{ID: uuid.New(), PostID: postID}}

	postRepo := &fakePostRepository{
		findPostByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dPost.Post, error) {
			return want, nil
		},
		listPostCommentsFunc: func(context.Context, database.Querier, uuid.UUID) ([]dPost.PostComment, error) {
			return wantComments, nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	got, err := svc.GetOnePost(context.Background(), postID)
	if err != nil {
		t.Fatalf("GetOnePost() error = %v, want nil", err)
	}
	if got.Post != want {
		t.Errorf("GetOnePost().Post = %+v, want %+v", got.Post, want)
	}
	if len(got.Comments) != 1 || got.Comments[0].ID != wantComments[0].ID {
		t.Errorf("GetOnePost().Comments = %+v, want %+v", got.Comments, wantComments)
	}
}

func TestService_GetManyPosts(t *testing.T) {
	want := []dPost.Post{{ID: uuid.New()}, {ID: uuid.New()}}

	postRepo := &fakePostRepository{
		listPostsFunc: func(context.Context, database.Querier, dPost.FilterQuery) ([]dPost.Post, error) {
			return want, nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	got, err := svc.GetManyPosts(context.Background(), dPost.FilterQuery{})
	if err != nil {
		t.Fatalf("GetManyPosts() error = %v, want nil", err)
	}
	if len(got) != len(want) {
		t.Errorf("GetManyPosts() returned %d posts, want %d", len(got), len(want))
	}
}

/* ==================== Tests: CreatePostComment ==================== */

func TestService_CreatePostComment(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()

	var added dPost.PostComment
	postRepo := &fakePostRepository{
		addCommentFunc: func(_ context.Context, _ database.Querier, c dPost.PostComment) error {
			added = c
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	cmd := PostCommentCommand{PostID: postID, Content: "Nice post!"}
	if _, err := svc.CreatePostComment(ctxWithUser(userID), cmd); err != nil {
		t.Fatalf("CreatePostComment() error = %v, want nil", err)
	}

	// Regression test: CreatePostComment used to overwrite comment.PostID
	// with the freshly generated comment ID instead of setting comment.ID,
	// which left every comment's own ID at its zero value and pointed
	// PostID at a random unrelated UUID instead of the actual post.
	if added.ID == uuid.Nil {
		t.Error("added.ID is zero, want a generated UUID")
	}
	if added.PostID != postID {
		t.Errorf("added.PostID = %v, want %v (the post being commented on)", added.PostID, postID)
	}
	if added.UserID != userID {
		t.Errorf("added.UserID = %v, want %v", added.UserID, userID)
	}
	if added.Content != "Nice post!" {
		t.Errorf("added.Content = %q, want %q", added.Content, "Nice post!")
	}
}

/* ==================== Tests: CreatePostEngagement ==================== */

func TestService_CreatePostEngagement(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()

	var added dPost.PostEngagement
	postRepo := &fakePostRepository{
		addEngagementFunc: func(_ context.Context, _ database.Querier, e dPost.PostEngagement) error {
			added = e
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	err := svc.CreatePostEngagement(ctxWithUser(userID), dPost.TypePost, dPost.ActionLike, postID)
	if err != nil {
		t.Fatalf("CreatePostEngagement() error = %v, want nil", err)
	}
	if added.ID == uuid.Nil {
		t.Error("added.ID is zero, want a generated UUID")
	}
	if added.AssetID != postID || added.AssetType != dPost.TypePost || added.UserID != userID || added.Type != dPost.ActionLike {
		t.Errorf("added = %+v, want AssetID=%v AssetType=POST UserID=%v Type=LIKE", added, postID, userID)
	}
}

/* ==================== Tests: verifyAccess ==================== */

func TestService_verifyAccess_MissingRequester(t *testing.T) {
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

	_, err := svc.CreatePost(context.Background(), PostCommand{Title: "x", Content: "y"})
	if !errors.Is(err, dUser.ErrInvalidUserID) {
		t.Errorf("CreatePost() error = %v, want %v", err, dUser.ErrInvalidUserID)
	}
}

/* ==================== Tests: replies, totals, authors ==================== */

func TestService_CreatePostComment_Reply(t *testing.T) {
	userID := uuid.New()
	postID := uuid.New()
	parentID := uuid.New()

	newSvc := func(parent dPost.PostComment, findErr error, added *dPost.PostComment) *Service {
		return NewService(fakeTransactor{}, &fakeUserRepository{}, &fakePostRepository{
			findCommentByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dPost.PostComment, error) {
				return parent, findErr
			},
			addCommentFunc: func(_ context.Context, _ database.Querier, c dPost.PostComment) error {
				*added = c
				return nil
			},
		}, &fakeNoteRepository{})
	}

	t.Run("stores the parent of a reply on the same post", func(t *testing.T) {
		var added dPost.PostComment
		svc := newSvc(dPost.PostComment{ID: parentID, PostID: postID}, nil, &added)

		cmd := PostCommentCommand{PostID: postID, Content: "Agreed", ParentID: &parentID}
		if _, err := svc.CreatePostComment(ctxWithUser(userID), cmd); err != nil {
			t.Fatalf("CreatePostComment() error = %v, want nil", err)
		}
		if added.ParentID == nil || *added.ParentID != parentID {
			t.Errorf("added.ParentID = %v, want %v", added.ParentID, parentID)
		}
	})

	t.Run("rejects a parent on another post", func(t *testing.T) {
		var added dPost.PostComment
		svc := newSvc(dPost.PostComment{ID: parentID, PostID: uuid.New()}, nil, &added)

		cmd := PostCommentCommand{PostID: postID, Content: "Agreed", ParentID: &parentID}
		if _, err := svc.CreatePostComment(ctxWithUser(userID), cmd); !errors.Is(err, dPost.ErrInvalidParentID) {
			t.Errorf("CreatePostComment() error = %v, want %v", err, dPost.ErrInvalidParentID)
		}
		if added.ID != uuid.Nil {
			t.Error("a comment was stored despite the invalid parent")
		}
	})

	t.Run("rejects a parent that does not exist", func(t *testing.T) {
		var added dPost.PostComment
		svc := newSvc(dPost.PostComment{}, dPost.ErrCommentNotFound, &added)

		cmd := PostCommentCommand{PostID: postID, Content: "Agreed", ParentID: &parentID}
		if _, err := svc.CreatePostComment(ctxWithUser(userID), cmd); !errors.Is(err, dPost.ErrInvalidParentID) {
			t.Errorf("CreatePostComment() error = %v, want %v", err, dPost.ErrInvalidParentID)
		}
	})
}

func TestService_CountPosts(t *testing.T) {
	search := "soil"
	var seen dPost.FilterQuery
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, &fakePostRepository{
		countPostsFunc: func(_ context.Context, _ database.Querier, f dPost.FilterQuery) (int, error) {
			seen = f
			return 7, nil
		},
	}, &fakeNoteRepository{})

	got, err := svc.CountPosts(context.Background(), dPost.FilterQuery{Search: &search})
	if err != nil {
		t.Fatalf("CountPosts() error = %v, want nil", err)
	}
	if got != 7 || seen.Search == nil || *seen.Search != search {
		t.Errorf("CountPosts() = %d with filter %+v, want 7 with search %q", got, seen, search)
	}
}

/* ==================== Tests: reactions ==================== */

func TestService_GetOnePost_ViewerReactions(t *testing.T) {
	viewerID := uuid.New()
	postID := uuid.New()
	liked := uuid.New()
	untouched := uuid.New()

	postRepo := &fakePostRepository{
		findPostByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dPost.Post, error) {
			return dPost.Post{ID: postID}, nil
		},
		listPostCommentsFunc: func(context.Context, database.Querier, uuid.UUID) ([]dPost.PostComment, error) {
			return []dPost.PostComment{{ID: liked, PostID: postID}, {ID: untouched, PostID: postID}}, nil
		},
		listReactionsFunc: func(_ context.Context, _ database.Querier, userID, _ uuid.UUID) (map[uuid.UUID]dPost.ActionType, error) {
			if userID != viewerID {
				t.Errorf("reactions asked for user %v, want %v", userID, viewerID)
			}
			return map[uuid.UUID]dPost.ActionType{postID: dPost.ActionDislike, liked: dPost.ActionLike}, nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	got, err := svc.GetOnePost(ctxWithUser(viewerID), postID)
	if err != nil {
		t.Fatalf("GetOnePost() error = %v, want nil", err)
	}
	if got.Post.ViewerReaction != dPost.ActionDislike {
		t.Errorf("Post.ViewerReaction = %q, want %q", got.Post.ViewerReaction, dPost.ActionDislike)
	}
	if got.Comments[0].ViewerReaction != dPost.ActionLike || got.Comments[1].ViewerReaction != "" {
		t.Errorf("comment reactions = %q and %q, want %q and none", got.Comments[0].ViewerReaction, got.Comments[1].ViewerReaction, dPost.ActionLike)
	}
}

func TestService_RemovePostEngagement(t *testing.T) {
	userID := uuid.New()
	commentID := uuid.New()

	var gotType dPost.PostType
	var gotAsset, gotUser uuid.UUID
	postRepo := &fakePostRepository{
		removeReactionFunc: func(_ context.Context, _ database.Querier, assetType dPost.PostType, assetID, user uuid.UUID) error {
			gotType, gotAsset, gotUser = assetType, assetID, user
			return nil
		},
	}
	svc := NewService(fakeTransactor{}, &fakeUserRepository{}, postRepo, &fakeNoteRepository{})

	if err := svc.RemovePostEngagement(ctxWithUser(userID), dPost.TypeComment, commentID); err != nil {
		t.Fatalf("RemovePostEngagement() error = %v, want nil", err)
	}
	if gotType != dPost.TypeComment || gotAsset != commentID || gotUser != userID {
		t.Errorf("RemoveReaction got %v %v %v, want %v %v %v", gotType, gotAsset, gotUser, dPost.TypeComment, commentID, userID)
	}

	if err := svc.RemovePostEngagement(context.Background(), dPost.TypeComment, commentID); err == nil {
		t.Error("RemovePostEngagement() without a signed-in user error = nil, want an error")
	}
}
