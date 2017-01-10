package wildcard_router

import (
	"net/http"
	"strings"
)

//
type WildcardRouter struct {
	middlewares []func(wirter http.ResponseWriter, request *http.Request)
	handlers    []http.Handler
}

// Initial
func New() *WildcardRouter {
	return &WildcardRouter{}
}

func (w *WildcardRouter) MountTo(mountTo string, mux *http.ServeMux) {
	mountTo = "/" + strings.Trim(mountTo, "/")
	mux.Handle(mountTo, w)
	mux.Handle(mountTo+"/", w)
}

func (w *WildcardRouter) AddHandler(handler http.Handler) {
	w.handlers = append(w.handlers, handler)
}

func (w *WildcardRouter) Use(middleware func(writer http.ResponseWriter, request *http.Request)) {
	w.middlewares = append(w.middlewares, middleware)
}

func (w *WildcardRouter) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	wildcardRouterWriter := &WildcardRouterWriter{writer, 0, false}
	for _, middleware := range w.middlewares {
		middleware(writer, req)
	}

	for _, handler := range w.handlers {
		if handler.ServeHTTP(wildcardRouterWriter, req); wildcardRouterWriter.isProcessed() {
			return
		}

		wildcardRouterWriter.reset()
	}

	wildcardRouterWriter.skipNotFoundCheck = true
	http.NotFound(wildcardRouterWriter, req)
}

//	WildcardRouterWriter will used to capture status
type WildcardRouterWriter struct {
	http.ResponseWriter
	status            int
	skipNotFoundCheck bool
}

func (w *WildcardRouterWriter) Status() int {
	return w.status
}

func (w *WildcardRouterWriter) WriteHeader(statusCode int) {
	if w.skipNotFoundCheck || statusCode != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(statusCode)
	}

	w.status = statusCode
}

func (w *WildcardRouterWriter) Write(date []byte) (int, error) {
	if w.skipNotFoundCheck || w.status != http.StatusNotFound {
		w.status = http.StatusOK
		return w.ResponseWriter.Write(date)
	}

	return 0, nil
}

func (w *WildcardRouterWriter) reset() {
	w.skipNotFoundCheck = false
	w.Header().Set("Content-Type", "")
	w.status = 0
}

func (w *WildcardRouterWriter) isProcessed() bool {
	return w.status != http.StatusNotFound && w.status != 0
}
