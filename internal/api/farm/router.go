package farm

import (
	"agropal/internal/api/middleware"
	"net/http"
)

type Router struct {
	Handler Handler
}

func NewRouter(
	mux *http.ServeMux,
	stack middleware.HandlerMiddleware,
	handler Handler,
) *Router {

	routes := []middleware.Routes{
		{
			Path:        "/farm",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOne,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
