package user

import (
	aUser "agropal/internal/application/user"
	"agropal/internal/domain/shared"
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

	var userIDPtr *uuid.UUID
	userID, err := h.IoHandler.IdentifierFromPathValue(r, "user_id")
	if err != nil {
		userIDPtr = nil
	} else {
		userIDPtr = &userID
	}

	user, err := h.UserService.GetByID(ctx, userIDPtr)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToUserResponse(user))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := h.IoHandler.IdentifierFromPathValue(r, "user_id")
	if err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.UserUpdateRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(userID)

	if err = h.UserService.Update(ctx, cmd); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var userIDPtr *uuid.UUID
	userID, err := h.IoHandler.IdentifierFromPathValue(r, "user_id")
	if err != nil {
		userIDPtr = nil
	} else {
		userIDPtr = &userID
	}

	query := r.URL.Query().Get("delete_type")
	var deleteType shared.DeleteType
	switch query {
	case string(shared.DeleteSoft):
		deleteType = shared.DeleteSoft
	case string(shared.DeleteHard):
		deleteType = shared.DeleteHard
	default:
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, errors.New("invalid delete type"))
		return
	}

	if err = h.UserService.Delete(ctx, deleteType, userIDPtr); err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
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
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, errors.New("invalid asset type"))
		return
	}

	assetIDStr := r.URL.Query().Get("asset_id")
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, errors.New("invalid asset ID"))
		return
	}

	users, err := h.UserService.GetUsers(ctx, assetType, assetID)
	if err != nil {
		h.IoHandler.WriteError(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	res := make([]response.UserResponse, len(users))
	for i, user := range users {
		res[i] = DomainToUserResponse(user)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, res)
}
