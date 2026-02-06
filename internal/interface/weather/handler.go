package weather

import (
	aWeather "agropal/internal/application/weather"
	"agropal/internal/request"
	"agropal/internal/response"
	"agropal/internal/utils"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	IoHandler      utils.IOHandler
	WeatherService aWeather.Service
}

func NewHandler(
	tool utils.IOHandler,
	weatherService aWeather.Service,
) Handler {
	return Handler{
		IoHandler:      tool,
		WeatherService: weatherService,
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.WeatherRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(farmID)

	if err = h.WeatherService.Create(ctx, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) GetByFarmID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	result, err := h.WeatherService.GetByFarmID(ctx, farmID)
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
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

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	var req request.WeatherRequest
	if err = h.IoHandler.ReadRequestBody(r, &req); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}
	cmd := req.ToCommand(farmID)

	if err = h.WeatherService.Update(ctx, farmID, cmd); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	farmID, err := getIdentifierFromPathValue(r, "farm_id")
	if err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	if err = h.WeatherService.Delete(ctx, farmID); err != nil {
		h.IoHandler.WriteError(ctx, w, response.AUTH_INVALID_CREDENTIALS, err)
		return
	}

	h.IoHandler.WriteSuccess(r.Context(), w, response.AUTH_INVALID_CREDENTIALS, nil)
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
