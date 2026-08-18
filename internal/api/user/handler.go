package user

import (
	aUser "agropal/internal/app/user"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	ioHandler   *utils.IOHandler
	userService *aUser.Service
}

func NewHandler(
	tool *utils.IOHandler,
	userService *aUser.Service,
) Handler {
	return Handler{
		ioHandler:   tool,
		userService: userService,
	}
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromPathValue(r)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
		return
	}

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToUserResponse(user))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromPathValue(r)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
		return
	}

	var req request.UserUpdateRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.userService.UpdateUser(ctx, req.ToCommand(userID)); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dUser.SuccessUpdated, nil)
}

// ChangePassword verifies the caller's current password before delegating
// to UserService.UpdateUserPassword, since that use case only accepts the
// new password and trusts the caller to have already authenticated it.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromPathValue(r)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
		return
	}

	var req request.ChangePasswordRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}
	if err = utils.ValidPassword(req.NewPassword); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	user, err := h.userService.GetByID(ctx, userID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Security.Password), []byte(req.OldPassword)); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrUnauthorized, nil)
		return
	}

	if err = h.userService.UpdateUserPassword(ctx, user.ID, req.NewPassword); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dUser.SuccessUpdated, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, err := userIDFromPathValue(r)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dUser.ErrInvalidUserID, nil)
		return
	}

	var deleteType shared.DeleteType
	switch r.URL.Query().Get("permanent") {
	case "true":
		deleteType = shared.DeleteHard
	case "false", "":
		deleteType = shared.DeleteSoft
	default:
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.userService.Delete(ctx, deleteType, userID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dUser.SuccessDeleted, nil)
}

func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var assetType shared.AssetType
	switch r.URL.Query().Get("asset_type") {
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
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	assetID, err := uuid.Parse(r.URL.Query().Get("asset_id"))
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	users, err := h.userService.GetUsers(ctx, assetType, assetID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	res := make([]response.UserResponse, len(users))
	for i, user := range users {
		res[i] = DomainToUserResponse(user)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, res)
}

// userIDFromPathValue treats the literal segment "me" as the caller
// themself: the app layer's Service resolves uuid.Nil to the requester's
// own ID.
func userIDFromPathValue(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.PathValue("user_id")
	if userIDStr == "" || userIDStr == "me" {
		return uuid.Nil, nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}

	return userID, nil
}
