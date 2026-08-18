package transaction

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
			Path:        "/farm/{farm_id}/product/{product_id}/transaction",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreateTransaction)},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/transaction",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyTransactions)},
		},
		{
			Path:        "/farm/{farm_id}/transaction",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyTransactions)},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/transaction/{transaction_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOne,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetOneTransaction)},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/transaction/{transaction_id}",
			Method:      http.MethodPatch,
			Handler:     handler.Update,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermUpdateTransaction)},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/transaction",
			Method:      http.MethodDelete,
			Handler:     handler.Delete,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteTransaction)},
		},
		{
			Path:        "/transaction",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteMany,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteManyTransactions)},
		},
		{
			Path:        "/transaction/all",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteAll,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermDeleteAllTransactions)},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
