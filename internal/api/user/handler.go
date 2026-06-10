package user

import (
	aUser "agropal/internal/application/user"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	IoHandler   utils.IOHandler
	UserService aUser.Service
}

func NewHandler(
	tool utils.IOHandler,
	userService aUser.Service,
) Handler {
	return Handler{
		IoHandler:   tool,
		UserService: userService,
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID uuid.UUID
	var err error

	userIDStr := r.PathValue("user_id")
	if userIDStr != "me" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			h.IoHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
			return
		}
	}

	user, err := h.UserService.GetByID(ctx, userID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToUserResponse(user))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID uuid.UUID
	var err error

	userIDStr := r.PathValue("user_id")
	if userIDStr != "me" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			h.IoHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
			return
		}
	}

	var req request.UserUpdateRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	cmd := req.ToCommand(userID)

	if err = h.UserService.Update(ctx, cmd); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userID uuid.UUID
	var err error

	userIDStr := r.PathValue("user_id")
	if userIDStr != "me" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			h.IoHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
			return
		}
	}

	query := r.URL.Query().Get("permanent")

	var deleteType shared.DeleteType
	switch query {
	case "true":
		deleteType = shared.DeleteHard
	case "false":
		deleteType = shared.DeleteSoft
	default:
		h.IoHandler.WriteError(r.Context(), w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.UserService.Delete(ctx, deleteType, userID); err != nil {
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var assetType shared.AssetType
	assetTypeStr := r.URL.Query().Get("asset_type")
	switch assetTypeStr {
	case string(shared.AssetUser):
		assetType = shared.AssetUser
	case string(shared.AssetFarm):
		assetType = shared.AssetFarm
	case string(shared.AssetProduct):
		assetType = shared.AssetProduct
	case string(shared.AssetPost):
		assetType = shared.AssetPost
	case string(shared.AssetComment):
		assetType = shared.AssetComment
	default:
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, errors.New("invalid asset type"))
		return
	}

	assetIDStr := r.URL.Query().Get("asset_id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, errors.New("invalid asset ResCode"))
		return
	}

	users, err := h.UserService.GetUsers(ctx, assetType, assetID)
	if err != nil {
		h.IoHandler.WriteError(r.Context(), w, dAuth.AUTH_INVALID_CREDENTIALS, nil)
		return
	}

	res := make([]response.UserResponse, len(users))
	for i, user := range users {
		res[i] = DomainToUserResponse(user)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, res)
}
