package post

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
			Path:        "/post",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/post",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/post/{post_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOne,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/post/{post_id}/comment",
			Method:      http.MethodPost,
			Handler:     handler.CreateComment,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/post/{post_id}/engagement",
			Method:      http.MethodPost,
			Handler:     handler.CreatePostEngagement,
			Middlewares: []middleware.HandlerMiddleware{},
		},
		{
			Path:        "/comment/{comment_id}/engagement",
			Method:      http.MethodPost,
			Handler:     handler.CreateCommentEngagement,
			Middlewares: []middleware.HandlerMiddleware{},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
