package location

import (
	aLocation "agropal/internal/app/location"
	dLocation "agropal/internal/domain/location"
	"agropal/internal/domain/shared"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"
)

type Handler struct {
	ioHandler       *utils.IOHandler
	locationService *aLocation.Service
}

func NewHandler(
	tool *utils.IOHandler,
	locationService *aLocation.Service,
) Handler {
	return Handler{
		ioHandler:       tool,
		locationService: locationService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidFarmID, nil)
		return
	}

	var req request.LocationRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	locationID, err := h.locationService.Create(ctx, farmID, req.ToCommand())
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dLocation.SuccessCreated, response.LocationCreatedResponse{LocationID: locationID})
}

func (h *Handler) GetByFarmID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidFarmID, nil)
		return
	}

	result, err := h.locationService.GetByFarmID(ctx, farmID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	locationResponse := make([]response.LocationResponse, len(result))
	for i, res := range result {
		locationResponse[i] = DomainToLocationResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, locationResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidFarmID, nil)
		return
	}

	locationID, err := h.ioHandler.IdentifierFromPathValue(r, "location_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidID, nil)
		return
	}

	var req request.LocationRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.locationService.Update(ctx, farmID, locationID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dLocation.SuccessUpdated, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidFarmID, nil)
		return
	}

	locationID, err := h.ioHandler.IdentifierFromPathValue(r, "location_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dLocation.ErrInvalidID, nil)
		return
	}

	if err = h.locationService.Delete(ctx, farmID, locationID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dLocation.SuccessDeleted, nil)
}
