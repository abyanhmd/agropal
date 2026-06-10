package farm

import (
	"agropal/internal/api/middleware"
	"agropal/internal/application/farm"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	IoHandler   utils.IOHandler
	FarmService farm.Service
}

func NewHandler(
	ioHandler utils.IOHandler,
	farmService farm.Service,
) Handler {
	return Handler{
		IoHandler:   ioHandler,
		FarmService: farmService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.FarmRequest
	if err := h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	userID, err := middleware.ExtractUserID(ctx)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(userID)

	if err = h.FarmService.Create(ctx, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	result, err := h.FarmService.GetOne(ctx, farmID)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToFarmResponse(result))
}

func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.FarmService.GetMany(ctx)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	farmResponse := make([]response.FarmResponse, len(result))
	for i, res := range result {
		farmResponse[i] = DomainToFarmResponse(res)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, farmResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.FarmRequest
	if err := h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(farmID)

	if err = h.FarmService.Update(ctx, farmID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	if err = h.FarmService.Delete(ctx, farmID); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func getIdentifierFromPathValue(r *http.Request, key string) (uuid.UUID, error) {
	idStr := r.PathValue(key)
	if idStr == "" {
		return uuid.Nil, nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
