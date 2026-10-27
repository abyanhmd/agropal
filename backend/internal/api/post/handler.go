package post

import (
	"agropal/internal/api/middleware"
	aPost "agropal/internal/app/post"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Handler struct {
	ioHandler   *utils.IOHandler
	postService *aPost.Service
}

func NewHandler(
	tool *utils.IOHandler,
	postService *aPost.Service,
) Handler {
	return Handler{
		ioHandler:   tool,
		postService: postService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.PostRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	post, err := h.postService.CreatePost(ctx, req.ToCommand(uuid.Nil))
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessCreated, response.PostCreatedResponse{PostID: post.ID, Slug: post.Slug})
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := h.ioHandler.IdentifierFromPathValue(r, "post_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidID, nil)
		return
	}

	result, err := h.postService.GetOnePost(ctx, postID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToPostDetailResponse(result))
}

func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter := filterFromQuery(r)
	if !dPost.ValidSort(filter.Sorting, dPost.PostSortColumns) {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidSort, nil)
		return
	}

	result, err := h.postService.GetManyPosts(ctx, filter)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	// The total ignores limit and offset, so a client can work out how many pages there are
	total, err := h.postService.CountPosts(ctx, filter)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}
	w.Header().Set(middleware.TotalCountHeader, strconv.Itoa(total))

	postResponse := make([]response.PostResponse, len(result))
	for i, res := range result {
		postResponse[i] = DomainToPostResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, postResponse)
}

func (h *Handler) GetAuthors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.postService.GetPostAuthors(ctx)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	authors := make([]response.PostAuthorResponse, len(result))
	for i, author := range result {
		authors[i] = response.PostAuthorResponse{UserID: author.UserID, Name: author.Name}
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, authors)
}

func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := h.ioHandler.IdentifierFromPathValue(r, "post_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidID, nil)
		return
	}

	var req request.PostCommentRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	commentID, err := h.postService.CreatePostComment(ctx, req.ToCommand(postID))
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessCommentCreated, response.CommentCreatedResponse{CommentID: commentID})
}

func (h *Handler) CreatePostEngagement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := h.ioHandler.IdentifierFromPathValue(r, "post_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidID, nil)
		return
	}

	h.createEngagement(w, r, ctx, dPost.TypePost, postID)
}

func (h *Handler) CreateCommentEngagement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commentID, err := h.ioHandler.IdentifierFromPathValue(r, "comment_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidCommentID, nil)
		return
	}

	h.createEngagement(w, r, ctx, dPost.TypeComment, commentID)
}

func (h *Handler) DeletePostEngagement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := h.ioHandler.IdentifierFromPathValue(r, "post_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidID, nil)
		return
	}

	h.deleteEngagement(w, ctx, dPost.TypePost, postID)
}

func (h *Handler) DeleteCommentEngagement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	commentID, err := h.ioHandler.IdentifierFromPathValue(r, "comment_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidCommentID, nil)
		return
	}

	h.deleteEngagement(w, ctx, dPost.TypeComment, commentID)
}

// deleteEngagement takes back the requester's like or dislike. Having none to take back is fine.
func (h *Handler) deleteEngagement(w http.ResponseWriter, ctx context.Context, assetType dPost.PostType, assetID uuid.UUID) {
	if err := h.postService.RemovePostEngagement(ctx, assetType, assetID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessEngagementDeleted, nil)
}

func (h *Handler) createEngagement(w http.ResponseWriter, r *http.Request, ctx context.Context, assetType dPost.PostType, assetID uuid.UUID) {
	var req request.PostEngagementRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	var actionType dPost.ActionType
	switch dPost.ActionType(req.Type) {
	case dPost.ActionLike, dPost.ActionDislike, dPost.ActionView:
		actionType = dPost.ActionType(req.Type)
	default:
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidActionType, nil)
		return
	}

	if err := h.postService.CreatePostEngagement(ctx, assetType, actionType, assetID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessEngagementCreated, nil)
}

func filterFromQuery(r *http.Request) dPost.FilterQuery {
	q := r.URL.Query()
	filter := dPost.FilterQuery{}

	if slug := q.Get("slug"); slug != "" {
		filter.Slug = &slug
	}
	if search := strings.TrimSpace(q.Get("q")); search != "" {
		filter.Search = &search
	}
	// user_id may be repeated to ask for several authors at once
	for _, userIDStr := range q["user_id"] {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserIDs = append(filter.UserIDs, userID)
		}
	}

	if limitStr, offsetStr := q.Get("limit"), q.Get("offset"); limitStr != "" || offsetStr != "" {
		pagination := shared.PaginationQuery{Limit: defaultPageSize}
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			pagination.Limit = min(limit, maxPageSize)
		}
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset > 0 {
			pagination.Offset = offset
		}
		filter.Pagination = &pagination
	}

	if col, dir := q.Get("sort_column"), q.Get("sort_dir"); col != "" || dir != "" {
		filter.Sorting = &shared.SortQuery{Column: col, Dir: dir}
	}

	return filter
}
