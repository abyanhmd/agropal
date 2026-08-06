package auth

import (
	aAuth "agropal/internal/app/auth"
	dAuth "agropal/internal/domain/auth"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/utils"
	"net/http"
)

type Handler struct {
	ioHandler   *utils.IOHandler
	authService *aAuth.Service
}

func NewHandler(
	ioHandler *utils.IOHandler,
	authService *aAuth.Service,
) *Handler {
	return &Handler{
		ioHandler:   ioHandler,
		authService: authService,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.UserRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	cmd, err := ToRegisterCommand(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	if err := h.authService.Register(ctx, cmd); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessRegister, nil)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.UserRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	cmd, err := ToLoginCommand(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	tokenPair, err := h.authService.Login(ctx, cmd)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, tokenPair)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("rid")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	tokenPair, err := h.authService.Refresh(ctx, cookie.Value)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, tokenPair)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie("rid")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err := h.authService.Invalidate(ctx, cookie.Value); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessLogout, nil)
}
