package product

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
			Path:        "/farm/{farm_id}/product",
			Method:      http.MethodPost,
			Handler:     handler.CreateProduct,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product",
			Method:      http.MethodGet,
			Handler:     handler.GetManyProducts,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOneProduct,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}",
			Method:      http.MethodPatch,
			Handler:     handler.UpdateProduct,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteProduct,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/price",
			Method:      http.MethodPost,
			Handler:     handler.CreateProductPrice,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/price/{price_id}",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteProductPrice,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/stock",
			Method:      http.MethodPost,
			Handler:     handler.CreateProductStock,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/farm/{farm_id}/product/{product_id}/stock/{stock_id}",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteProductStock,
			Middlewares: []middleware.HandlerMiddleware{},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
