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
	IoHandler      *utils.IOHandler
	WeatherService *aWeather.Service
}

func NewHandler(
	tool *utils.IOHandler,
	weatherService *aWeather.Service,
) Handler {
	return Handler{
		IoHandler:      tool,
		WeatherService: weatherService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	var req request.WeatherRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.WeatherService.Create(ctx, farmID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dWeather.SuccessCreated, nil)
}

func (h *Handler) GetByFarmID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	result, err := h.WeatherService.GetByFarmID(ctx, farmID)
	if err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	weatherResponse := make([]response.WeatherResponse, len(result))
	for i, res := range result {
		weatherResponse[i] = DomainToWeatherResponse(res)
	}

	h.IoHandler.WriteResponseBody(w, http.StatusOK, weatherResponse)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	weatherID, err := h.IoHandler.IdentifierFromPathValue(r, "weather_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidID, nil)
		return
	}

	var req request.WeatherRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, shared.ErrInvalidRequest, nil)
		return
	}

	if err = h.WeatherService.Update(ctx, farmID, weatherID, req.ToCommand()); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dWeather.SuccessUpdated, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := h.IoHandler.IdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidFarmID, nil)
		return
	}

	weatherID, err := h.IoHandler.IdentifierFromPathValue(r, "weather_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, dWeather.ErrInvalidID, nil)
		return
	}

	if err = h.WeatherService.Delete(ctx, farmID, weatherID); err != nil {
		h.IoHandler.HandleServiceError(ctx, w, err)
		return
	}

	h.IoHandler.WriteSuccess(ctx, w, dWeather.SuccessDeleted, nil)
}
