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
	ioHandler   *utils.IOHandler
	farmService *aFarm.Service
}

func NewHandler(
	ioHandler *utils.IOHandler,
	farmService *aFarm.Service,
) Handler {
	return Handler{
		ioHandler:   ioHandler,
		farmService: farmService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req request.FarmRequest
	if err := h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	farmID, err := h.farmService.Create(ctx, req.ToCommand())
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dFarm.SuccessCreated, response.FarmCreatedResponse{FarmID: farmID})
}

func (h *Handler) GetOne(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	result, err := h.farmService.GetOne(ctx, farmID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, DomainToFarmResponse(result))
}

func (h *Handler) GetMany(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.farmService.GetMany(ctx)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	farmResponse := make([]response.FarmResponse, len(result))
	for i, res := range result {
		farmResponse[i] = DomainToFarmResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, farmResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	var req request.FarmRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.farmService.Update(ctx, farmID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dFarm.SuccessUpdated, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dFarm.ErrInvalidID, nil)
		return
	}

	if err = h.farmService.Delete(ctx, farmID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dFarm.SuccessDeleted, nil)
}
