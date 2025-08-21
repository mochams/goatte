// Interfaces and types definitions used in this package.

package goatte

import "net/http"

// Router handles HTTP requests by mapping URLs to functions. It supports middleware,
// namespaces for organizing routes, and mounting sub-routers under prefixes.
type Router struct {
	mux             *http.ServeMux
	middlewareFuncs []Middleware
	namespace       string
}

// Middleware is a function that adds extra features to a request handler, like
// logging requests.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// View is an interface for handling multiple HTTP methods (like GET or POST) for a
// single URL. Implement this to define how a URL responds to different request types.
type View interface {
	Head(w http.ResponseWriter, r *http.Request)
	Options(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	Patch(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
	Put(w http.ResponseWriter, r *http.Request)
}
