package post

import (
	"agropal/config"
	aPost "agropal/internal/app/post"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"agropal/internal/response"
	"agropal/internal/utils"
	"agropal/pkg"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

/* ==================== Fakes ====================== */

// fakeTransactor runs WithTx's callback directly against a nil Querier; the
// fake repositories below never touch the Querier they're given.
type fakeTransactor struct{}

func (fakeTransactor) DB() database.Querier { return nil }

func (fakeTransactor) WithTx(_ context.Context, fn func(q database.Querier) error) error {
	return fn(nil)
}

// fakeUserRepository implements dUser.Repository. Only FindUserByID is
// exercised by post.Service (through verifyAccess in CreatePost).
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

// fakePostRepository implements dPost.Repository. Only the methods the
// handler tests exercise are wired to a *Func field; everything else panics
// if called.
type fakePostRepository struct {
	createPostFunc       func(ctx context.Context, q database.Querier, p dPost.Post) error
	findPostByIDFunc     func(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error)
	listPostsFunc        func(ctx context.Context, q database.Querier, filter dPost.FilterQuery) ([]dPost.Post, error)
	addCommentFunc       func(ctx context.Context, q database.Querier, c dPost.PostComment) error
	countPostsFunc       func(ctx context.Context, q database.Querier, filter dPost.FilterQuery) (int, error)
	findCommentByIDFunc  func(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.PostComment, error)
	listPostCommentsFunc func(ctx context.Context, q database.Querier, postID uuid.UUID) ([]dPost.PostComment, error)
	addEngagementFunc    func(ctx context.Context, q database.Querier, e dPost.PostEngagement) error
	removeReactionFunc   func(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID, userID uuid.UUID) error
	listReactionsFunc    func(ctx context.Context, q database.Querier, userID, postID uuid.UUID) (map[uuid.UUID]dPost.ActionType, error)
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

// newTestHandler wires a real *aPost.Service (backed by fakes) into a real
// Handler, so these tests exercise the handler's actual request parsing,
// status codes, and response bodies rather than a mocked-out service.
func newTestHandler(userRepo dUser.Repository, postRepo dPost.Repository, noteRepo dNote.Repository) Handler {
	logger := pkg.NewLogger(config.Config{})
	svc := aPost.NewService(fakeTransactor{}, userRepo, postRepo, noteRepo)
	ioHandler := utils.NewIOHandler(logger, nil, utils.NewValidator())

	return NewHandler(ioHandler, svc)
}

func ctxWithUser(userID uuid.UUID) context.Context {
	return shared.WithMetadata(context.Background(), shared.KeyUserID, userID.String())
}

func jsonRequestWithContext(ctx context.Context, method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	return req.WithContext(ctx)
}

func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) utils.StandardEnvelope {
	t.Helper()

	var envelope utils.StandardEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
	}
	return envelope
}

func withPathValue(r *http.Request, key, value string) *http.Request {
	r.SetPathValue(key, value)
	return r
}

/* ==================== Tests: Create ====================== */

func TestHandler_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				return dUser.User{ID: userID, Account: dUser.AccountInfo{Name: "Jane"}}, nil
			},
		}
		postRepo := &fakePostRepository{
			createPostFunc: func(context.Context, database.Querier, dPost.Post) error { return nil },
		}
		noteRepo := &fakeNoteRepository{
			createNoteFunc: func(context.Context, database.Querier, dNote.Note) error { return nil },
		}
		h := newTestHandler(userRepo, postRepo, noteRepo)

		req := jsonRequestWithContext(ctxWithUser(userID), http.MethodPost, "/post", `{"title":"Hello","content":"World"}`)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.SuccessCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.SuccessCreated.String())
		}
		data, _ := envelope.Data.(map[string]any)
		if id, _ := data["post_id"].(string); id == "" {
			t.Errorf("data.post_id is empty, want the new post's id (data=%v)", envelope.Data)
		}
		if slug, _ := data["slug"].(string); !strings.HasPrefix(slug, "hello-") {
			t.Errorf("data.slug = %q, want a slug built from the title", slug)
		}
	})

	t.Run("Invalid JSON Body", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post", "not-json")
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrInvalidRequest.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrInvalidRequest.Error())
		}
	})

	t.Run("Missing Required Field Fails Validation Before Hitting The Service", func(t *testing.T) {
		userRepo := &fakeUserRepository{
			findUserByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dUser.User, error) {
				t.Fatal("service should not be called when the request fails validation")
				return dUser.User{}, nil
			},
		}
		h := newTestHandler(userRepo, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post", `{"content":"World"}`)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != shared.ErrValidation.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, shared.ErrValidation.Error())
		}
	})

	t.Run("Missing Authenticated User Maps To An Error", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		// No shared.KeyUserID in context, simulating a request that somehow
		// reached the handler without going through the auth middleware.
		req := jsonRequestWithContext(context.Background(), http.MethodPost, "/post", `{"title":"Hello","content":"World"}`)
		rec := httptest.NewRecorder()

		h.Create(rec, req)

		if rec.Code == http.StatusCreated {
			t.Fatalf("status = %d, want a non-201 status when no authenticated user is present", rec.Code)
		}
	})
}

/* ==================== Tests: GetOne ====================== */

func TestHandler_GetOne(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		postID := uuid.New()
		postRepo := &fakePostRepository{
			findPostByIDFunc: func(_ context.Context, _ database.Querier, id uuid.UUID) (dPost.Post, error) {
				return dPost.Post{ID: id, Title: "Hello", Content: "World"}, nil
			},
			listPostCommentsFunc: func(context.Context, database.Querier, uuid.UUID) ([]dPost.PostComment, error) {
				return nil, nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := withPathValue(httptest.NewRequest(http.MethodGet, "/post/"+postID.String(), nil), "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got response.PostDetailResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if got.Post.PostID != postID {
			t.Errorf("Post.PostID = %v, want %v", got.Post.PostID, postID)
		}
	})

	t.Run("Invalid Post ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := withPathValue(httptest.NewRequest(http.MethodGet, "/post/not-a-uuid", nil), "post_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.ErrInvalidID.Error())
		}
	})

	t.Run("Not Found Maps To 404", func(t *testing.T) {
		postRepo := &fakePostRepository{
			findPostByIDFunc: func(context.Context, database.Querier, uuid.UUID) (dPost.Post, error) {
				return dPost.Post{}, dPost.ErrNotFound
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		postID := uuid.New()
		req := withPathValue(httptest.NewRequest(http.MethodGet, "/post/"+postID.String(), nil), "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.GetOne(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}

/* ==================== Tests: GetMany ====================== */

func TestHandler_GetMany(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		postID := uuid.New()
		postRepo := &fakePostRepository{
			listPostsFunc: func(context.Context, database.Querier, dPost.FilterQuery) ([]dPost.Post, error) {
				return []dPost.Post{{ID: postID, Title: "Hello", AuthorName: "Jane"}}, nil
			},
			countPostsFunc: func(context.Context, database.Querier, dPost.FilterQuery) (int, error) { return 42, nil },
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := httptest.NewRequest(http.MethodGet, "/post", nil)
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if got := rec.Header().Get("X-Total-Count"); got != "42" {
			t.Errorf("X-Total-Count = %q, want %q", got, "42")
		}
		var got []response.PostResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body %q: %v", rec.Body.String(), err)
		}
		if len(got) != 1 || got[0].PostID != postID {
			t.Errorf("got %+v, want a single post with ID %v", got, postID)
		}
		if got[0].AuthorName != "Jane" {
			t.Errorf("AuthorName = %q, want %q", got[0].AuthorName, "Jane")
		}
	})

	t.Run("Rejects A Sort Outside The Allowed Columns", func(t *testing.T) {
		for _, query := range []string{
			"sort_column=created_at;DROP TABLE posts&sort_dir=asc",
			"sort_column=password&sort_dir=asc",
			"sort_column=created_at&sort_dir=sideways",
			"sort_column=created_at",
		} {
			h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

			req := httptest.NewRequest(http.MethodGet, "/post", nil)
			req.URL.RawQuery = query
			rec := httptest.NewRecorder()

			h.GetMany(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("%q: status = %d, want %d (body=%s)", query, rec.Code, http.StatusBadRequest, rec.Body.String())
			}
		}
	})

	t.Run("Passes Filters, Search And A Clamped Page To The Repository", func(t *testing.T) {
		authorA, authorB := uuid.New(), uuid.New()
		var listed, counted dPost.FilterQuery
		postRepo := &fakePostRepository{
			listPostsFunc: func(_ context.Context, _ database.Querier, f dPost.FilterQuery) ([]dPost.Post, error) {
				listed = f
				return nil, nil
			},
			countPostsFunc: func(_ context.Context, _ database.Querier, f dPost.FilterQuery) (int, error) {
				counted = f
				return 0, nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := httptest.NewRequest(http.MethodGet, "/post", nil)
		req.URL.RawQuery = "user_id=" + authorA.String() + "&user_id=" + authorB.String() + "&q=%20soil%20&limit=9999&offset=-5&sort_column=like_count&sort_dir=desc"
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if len(listed.UserIDs) != 2 || listed.UserIDs[0] != authorA || listed.UserIDs[1] != authorB {
			t.Errorf("UserIDs = %v, want [%v %v]", listed.UserIDs, authorA, authorB)
		}
		if listed.Search == nil || *listed.Search != "soil" {
			t.Errorf("Search = %v, want %q", listed.Search, "soil")
		}
		if listed.Pagination == nil || listed.Pagination.Limit != 100 || listed.Pagination.Offset != 0 {
			t.Errorf("Pagination = %+v, want limit 100 and offset 0", listed.Pagination)
		}
		if listed.Sorting == nil || listed.Sorting.Column != "like_count" || listed.Sorting.Dir != "desc" {
			t.Errorf("Sorting = %+v, want like_count desc", listed.Sorting)
		}
		if counted.Search == nil || len(counted.UserIDs) != 2 {
			t.Errorf("count filter = %+v, want the same search and authors as the list", counted)
		}
	})

	t.Run("Service Error Maps To 500", func(t *testing.T) {
		postRepo := &fakePostRepository{
			listPostsFunc: func(context.Context, database.Querier, dPost.FilterQuery) ([]dPost.Post, error) {
				return nil, dPost.ErrCreateFailed
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := httptest.NewRequest(http.MethodGet, "/post", nil)
		rec := httptest.NewRecorder()

		h.GetMany(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusInternalServerError, rec.Body.String())
		}
	})
}

/* ==================== Tests: CreateComment ====================== */

func TestHandler_CreateComment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		postID := uuid.New()
		postRepo := &fakePostRepository{
			addCommentFunc: func(_ context.Context, _ database.Querier, c dPost.PostComment) error {
				if c.PostID != postID {
					t.Errorf("AddComment PostID = %v, want %v", c.PostID, postID)
				}
				if c.UserID != userID {
					t.Errorf("AddComment UserID = %v, want %v", c.UserID, userID)
				}
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(userID), http.MethodPost, "/post/"+postID.String()+"/comment", `{"content":"Nice post!"}`)
		req = withPathValue(req, "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.CreateComment(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.SuccessCommentCreated.String() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.SuccessCommentCreated.String())
		}
	})

	t.Run("Invalid Post ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post/not-a-uuid/comment", `{"content":"Nice post!"}`)
		req = withPathValue(req, "post_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.CreateComment(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.ErrInvalidID.Error())
		}
	})

	t.Run("Missing Required Field Fails Validation", func(t *testing.T) {
		postID := uuid.New()
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post/"+postID.String()+"/comment", `{}`)
		req = withPathValue(req, "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.CreateComment(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

/* ==================== Tests: Engagement ====================== */

func TestHandler_CreatePostEngagement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		postID := uuid.New()
		postRepo := &fakePostRepository{
			addEngagementFunc: func(_ context.Context, _ database.Querier, e dPost.PostEngagement) error {
				if e.AssetID != postID || e.AssetType != dPost.TypePost || e.Type != dPost.ActionLike || e.UserID != userID {
					t.Errorf("AddEngagement got %+v, want AssetID=%v AssetType=%v Type=%v UserID=%v", e, postID, dPost.TypePost, dPost.ActionLike, userID)
				}
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(userID), http.MethodPost, "/post/"+postID.String()+"/engagement", `{"type":"LIKE"}`)
		req = withPathValue(req, "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.CreatePostEngagement(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("Invalid Action Type Maps To 400", func(t *testing.T) {
		postID := uuid.New()
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post/"+postID.String()+"/engagement", `{"type":"NOT_A_TYPE"}`)
		req = withPathValue(req, "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.CreatePostEngagement(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.ErrInvalidActionType.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.ErrInvalidActionType.Error())
		}
	})

	t.Run("Invalid Post ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/post/not-a-uuid/engagement", `{"type":"LIKE"}`)
		req = withPathValue(req, "post_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.CreatePostEngagement(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.ErrInvalidID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.ErrInvalidID.Error())
		}
	})
}

func TestHandler_CreateCommentEngagement(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		userID := uuid.New()
		commentID := uuid.New()
		postRepo := &fakePostRepository{
			addEngagementFunc: func(_ context.Context, _ database.Querier, e dPost.PostEngagement) error {
				if e.AssetID != commentID || e.AssetType != dPost.TypeComment || e.Type != dPost.ActionView {
					t.Errorf("AddEngagement got %+v, want AssetID=%v AssetType=%v Type=%v", e, commentID, dPost.TypeComment, dPost.ActionView)
				}
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(userID), http.MethodPost, "/comment/"+commentID.String()+"/engagement", `{"type":"VIEW"}`)
		req = withPathValue(req, "comment_id", commentID.String())
		rec := httptest.NewRecorder()

		h.CreateCommentEngagement(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusCreated, rec.Body.String())
		}
	})

	t.Run("Invalid Comment ID Maps To 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		req := jsonRequestWithContext(ctxWithUser(uuid.New()), http.MethodPost, "/comment/not-a-uuid/engagement", `{"type":"VIEW"}`)
		req = withPathValue(req, "comment_id", "not-a-uuid")
		rec := httptest.NewRecorder()

		h.CreateCommentEngagement(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		envelope := decodeEnvelope(t, rec)
		if envelope.Info.Key != dPost.ErrInvalidCommentID.Error() {
			t.Errorf("Info.Key = %q, want %q", envelope.Info.Key, dPost.ErrInvalidCommentID.Error())
		}
	})
}

/* ==================== Tests: DeleteEngagement ====================== */

func TestHandler_DeleteEngagement(t *testing.T) {
	userID := uuid.New()

	t.Run("Post reaction is taken back", func(t *testing.T) {
		postID := uuid.New()
		called := false
		postRepo := &fakePostRepository{
			removeReactionFunc: func(_ context.Context, _ database.Querier, assetType dPost.PostType, assetID, user uuid.UUID) error {
				called = true
				if assetType != dPost.TypePost || assetID != postID || user != userID {
					t.Errorf("RemoveReaction got %v %v %v, want %v %v %v", assetType, assetID, user, dPost.TypePost, postID, userID)
				}
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := withPathValue(jsonRequestWithContext(ctxWithUser(userID), http.MethodDelete, "/post/"+postID.String()+"/engagement", ""), "post_id", postID.String())
		rec := httptest.NewRecorder()

		h.DeletePostEngagement(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
		if key := decodeEnvelope(t, rec).Info.Key; key != dPost.SuccessEngagementDeleted.String() {
			t.Errorf("Info.Key = %q, want %q", key, dPost.SuccessEngagementDeleted.String())
		}
		if !called {
			t.Error("the repository was not asked to remove the reaction")
		}
	})

	t.Run("Comment reaction is taken back", func(t *testing.T) {
		commentID := uuid.New()
		postRepo := &fakePostRepository{
			removeReactionFunc: func(_ context.Context, _ database.Querier, assetType dPost.PostType, assetID, _ uuid.UUID) error {
				if assetType != dPost.TypeComment || assetID != commentID {
					t.Errorf("RemoveReaction got %v %v, want %v %v", assetType, assetID, dPost.TypeComment, commentID)
				}
				return nil
			},
		}
		h := newTestHandler(&fakeUserRepository{}, postRepo, &fakeNoteRepository{})

		req := withPathValue(jsonRequestWithContext(ctxWithUser(userID), http.MethodDelete, "/comment/"+commentID.String()+"/engagement", ""), "comment_id", commentID.String())
		rec := httptest.NewRecorder()

		h.DeleteCommentEngagement(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (body=%s)", rec.Code, http.StatusOK, rec.Body.String())
		}
	})

	t.Run("Invalid ids map to 400", func(t *testing.T) {
		h := newTestHandler(&fakeUserRepository{}, &fakePostRepository{}, &fakeNoteRepository{})

		rec := httptest.NewRecorder()
		h.DeletePostEngagement(rec, withPathValue(jsonRequestWithContext(ctxWithUser(userID), http.MethodDelete, "/post/x/engagement", ""), "post_id", "x"))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("post: status = %d, want %d", rec.Code, http.StatusBadRequest)
		}

		rec = httptest.NewRecorder()
		h.DeleteCommentEngagement(rec, withPathValue(jsonRequestWithContext(ctxWithUser(userID), http.MethodDelete, "/comment/x/engagement", ""), "comment_id", "x"))
		if rec.Code != http.StatusBadRequest {
			t.Errorf("comment: status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
