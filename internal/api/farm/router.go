package farm

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
			Path:        "/farm",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreateFarm)},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOne,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetOneFarm)},
		},
		{
			Path:        "/farm",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyFarms)},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermUpdateFarm)},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteFarm)},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
