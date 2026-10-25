package middleware

import "net/http"

type Routes struct {
	Path        string
	Method      string
	Handler     http.HandlerFunc
	Middlewares []HandlerMiddleware
}

func SetupV1Routes(routes []Routes, stack HandlerMiddleware, mux *http.ServeMux) {
	routeMap := make(map[string]map[string]http.Handler)

	for _, route := range routes {
		final := Chain(route.Handler, route.Middlewares...)
		final = stack(final)

		if _, ok := routeMap[route.Path]; !ok {
			routeMap[route.Path] = make(map[string]http.Handler)
		}
		routeMap[route.Path][route.Method] = final
	}

	for path, methods := range routeMap {
		routePrefix := "/api/v1"

		mux.HandleFunc(routePrefix+path, func(w http.ResponseWriter, r *http.Request) {
			if handler, ok := methods[r.Method]; ok {
				handler.ServeHTTP(w, r)
				return
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
		})
	}
}
