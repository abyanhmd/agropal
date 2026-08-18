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

func (h *Handler) VerifyRegister(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.VerificationRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	email, code, err := ToVerification(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	if err := h.authService.VerifyRegistration(ctx, email, code); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessEmailVerified, nil)
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

	if err := h.authService.Login(ctx, cmd); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessVerificationSent, nil)
}

func (h *Handler) VerifyLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.VerificationRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	email, code, err := ToVerification(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	tokenPair, err := h.authService.VerifyLogin(ctx, email, code)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessLogin, tokenPair)
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

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessRefresh, tokenPair)
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

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.PasswordResetRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	email, err := ToPasswordResetEmail(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	if err := h.authService.RequestPasswordReset(ctx, email); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessPasswordResetEmail, nil)
}

func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.PasswordResetConfirmRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	email, code, newPassword, err := ToPasswordResetConfirmation(req)
	if err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	if err := h.authService.ConfirmPasswordReset(ctx, email, code, newPassword); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrorKey(err.Error()), nil)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dAuth.SuccessPasswordReset, nil)
}
