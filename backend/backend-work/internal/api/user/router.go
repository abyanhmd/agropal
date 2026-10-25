package user

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
			Path:        "/user/{user_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetByID,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetUserByID)},
		},
		{
			Path:        "/user/{user_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermUpdateUser)},
		},
		{
			Path:        "/user/{user_id}/password",
			Method:      http.MethodPatch,
			Handler:     handler.ChangePassword,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermChangeUserPassword)},
		},
		{
			Path:        "/user/{user_id}",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteUser)},
		},
		{
			Path:        "/users",
			Method:      http.MethodGet,
			Handler:     handler.GetUsers,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyUsers)},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
