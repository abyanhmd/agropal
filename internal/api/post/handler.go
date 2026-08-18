package post

import (
	aPost "agropal/internal/app/post"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
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

	if err := h.postService.CreatePost(ctx, req.ToCommand(uuid.Nil)); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessCreated, nil)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	postID, err := h.ioHandler.IdentifierFromPathValue(r, "post_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dPost.ErrInvalidID, nil)
		return
	}

	result, err := h.postService.GetOnePost(ctx, postID, filterFromQuery(r))
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToPostDetailResponse(result))
}

func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.postService.GetManyPosts(ctx, filterFromQuery(r))
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	postResponse := make([]response.PostResponse, len(result))
	for i, res := range result {
		postResponse[i] = DomainToPostResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, postResponse)
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

	if err = h.postService.CreatePostComment(ctx, req.ToCommand(postID)); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dPost.SuccessCommentCreated, nil)
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

	h.ioHandler.WriteResponseBody(w, http.StatusCreated, nil)
}

func filterFromQuery(r *http.Request) dPost.FilterQuery {
	q := r.URL.Query()
	filter := dPost.FilterQuery{}

	if slug := q.Get("slug"); slug != "" {
		filter.Slug = &slug
	}
	if userIDStr := q.Get("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}

	if limitStr, offsetStr := q.Get("limit"), q.Get("offset"); limitStr != "" || offsetStr != "" {
		pagination := shared.PaginationQuery{}
		if limit, err := strconv.Atoi(limitStr); err == nil {
			pagination.Limit = limit
		}
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			pagination.Offset = offset
		}
		filter.Pagination = &pagination
	}

	if col, dir := q.Get("sort_column"), q.Get("sort_dir"); col != "" || dir != "" {
		filter.Sorting = &shared.SortQuery{Column: col, Dir: dir}
	}

	return filter
}
