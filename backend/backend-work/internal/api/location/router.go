package location

import (
	"agropal/config"
	"agropal/internal/api/middleware"
	"net/http"
)

type Router struct {
	Handler Handler
}

func NewRouter(
	mux *http.ServeMux,
	stack middleware.HandlerMiddleware,
	mw *middleware.Middleware,
	handler Handler,
) *Router {

	routes := []middleware.Routes{
		{
			Path:        "/farm/{farm_id}/location",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreateLocation)},
		},
		{
			Path:        "/farm/{farm_id}/location",
			Method:      http.MethodGet,
			Handler:     handler.GetByFarmID,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetLocationByFarmID)},
		},
		{
			Path:        "/farm/{farm_id}/location/{location_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermUpdateLocation)},
		},
		{
			Path:        "/farm/{farm_id}/location/{location_id}",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteLocation)},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
