package user

import (
	"agropal/internal/interface/middleware"
	"net/http"
)

type Router struct {
	Handler Handler
}

func NewRouter(
	mux *http.ServeMux,
	m *middleware.Middleware,
	stack middleware.HandlerMiddleware,
	handler Handler,
) *Router {

	routes := []middleware.Routes{
		{
			Path:        "/user/{user_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetByID,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/user/{user_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/user/{user_id}",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/users",
			Method:      http.MethodGet,
			Handler:     handler.GetUsers,
			Middlewares: []middleware.HandlerMiddleware{},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
