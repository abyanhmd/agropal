package weather

import (
	aWeather "agropal/internal/app/weather"
	"agropal/internal/domain/shared"
	dWeather "agropal/internal/domain/weather"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"
)

type Handler struct {
	ioHandler      *utils.IOHandler
	weatherService *aWeather.Service
}

func NewHandler(
	tool *utils.IOHandler,
	weatherService *aWeather.Service,
) Handler {
	return Handler{
		ioHandler:      tool,
		weatherService: weatherService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	var req request.WeatherRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.weatherService.Create(ctx, farmID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dWeather.SuccessCreated, nil)
}

func (h *Handler) GetByFarmID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	result, err := h.weatherService.GetByFarmID(ctx, farmID)
	if err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	weatherResponse := make([]response.WeatherResponse, len(result))
	for i, res := range result {
		weatherResponse[i] = DomainToWeatherResponse(res)
	}

	h.ioHandler.WriteResponseBody(w, http.StatusOK, weatherResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	weatherID, err := h.ioHandler.IdentifierFromPathValue(r, "weather_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidID, nil)
		return
	}

	var req request.WeatherRequest
	if err = h.ioHandler.ReadRequestBody(r, &req); err != nil {
		h.ioHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}
	if !h.ioHandler.ValidateStruct(ctx, w, req) {
		return
	}

	if err = h.weatherService.Update(ctx, farmID, weatherID, req.ToCommand()); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dWeather.SuccessUpdated, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.ioHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	weatherID, err := h.ioHandler.IdentifierFromPathValue(r, "weather_id")
	if err != nil {
		h.ioHandler.WriteError(ctx, w, dWeather.ErrInvalidID, nil)
		return
	}

	if err = h.weatherService.Delete(ctx, farmID, weatherID); err != nil {
		h.ioHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.ioHandler.WriteSuccess(ctx, w, dWeather.SuccessDeleted, nil)
}
