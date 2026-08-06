package farm

import (
	aFarm "agropal/internal/app/farm"
	dFarm "agropal/internal/domain/farm"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"
)

type Handler struct {
	IoHandler   *utils.IOHandler
	FarmService *aFarm.Service
}

func NewHandler(
	ioHandler *utils.IOHandler,
	farmService *aFarm.Service,
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
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if req.Name == "" {
		h.IoHandler.WriteError(ctx, w, dFarm.ErrInvalidName, nil)
		return
	}

	if err := h.FarmService.Create(ctx, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusCreated, nil)
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	result, err := h.FarmService.GetOne(ctx, farmID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, DomainToFarmResponse(result))
}

func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.FarmService.GetMany(ctx)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
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

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	var req request.FarmRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.FarmService.Update(ctx, farmID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	if err = h.FarmService.Delete(ctx, farmID); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteResponseBody(w, http.StatusNoContent, nil)
}
