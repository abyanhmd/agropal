package post

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
			Path:        "/post",
			Method:      http.MethodPost,
			Handler:     handler.Create,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreatePost)},
		},
		{
			Path:        "/post",
			Method:      http.MethodGet,
			Handler:     handler.GetMany,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyPosts)},
		},
		{
			Path:        "/post/authors",
			Method:      http.MethodGet,
			Handler:     handler.GetAuthors,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetManyPosts)},
		},
		{
			Path:        "/post/{post_id}",
			Method:      http.MethodGet,
			Handler:     handler.GetOne,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermGetOnePost)},
		},
		{
			Path:        "/post/{post_id}/comment",
			Method:      http.MethodPost,
			Handler:     handler.CreateComment,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreatePostComment)},
		},
		{
			Path:        "/post/{post_id}/engagement",
			Method:      http.MethodPost,
			Handler:     handler.CreatePostEngagement,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreatePostEngagement)},
		},
		{
			Path:        "/post/{post_id}/engagement",
			Method:      http.MethodDelete,
			Handler:     handler.DeletePostEngagement,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreatePostEngagement)},
		},
		{
			Path:        "/comment/{comment_id}/engagement",
			Method:      http.MethodDelete,
			Handler:     handler.DeleteCommentEngagement,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreateCommentEngagement)},
		},
		{
			Path:        "/comment/{comment_id}/engagement",
			Method:      http.MethodPost,
			Handler:     handler.CreateCommentEngagement,
			Middlewares: []middleware.HandlerMiddleware{mw.RequirePermission(config.PermCreateCommentEngagement)},
		},
	}

	middleware.SetupV1Routes(routes, stack, mux)

	return &Router{
		Handler: handler,
	}
}
