// A lightweight HTTP router built on top of the standard net/http package.
// It supports routing for HTTP methods, middleware, static file serving,
// and mounting sub-routers with namespace support.

package goatte

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// routeRegistry maps namespace:name pairs to their corresponding URL paths for URL reversing.
var routeRegistry = make(map[string]string)

// pathParamRegex is a pattern used to find placeholders in URL paths, like "{name}" in
// "/users/{name}/".
var pathParamRegex = regexp.MustCompile(`\{(\w+)\}`)

// MethodAll is used to indicate method-agnostic routing (matches MethodAll HTTP methods)
const MethodAll = "ALL"

// NewRouter creates a new Router with a namespace and optional middleware.
// The namespace (e.g., "api" or "web") helps organize routes and acts as path prefix.
// Middleware functions are applied to all routes in this router.
//
// Example:
//
//	router := NewRouter("api", loggerMiddleware, authMiddleware)
//	router.Get("/users", "users-list", userHandler)
func NewRouter(namespace string, middlewareFuncs ...Middleware) *Router {
	return &Router{
		mux:             http.NewServeMux(),
		middlewareFuncs: middlewareFuncs,
		namespace:       namespace,
	}
}

// ServeHTTP implements the http.Handler interface, forwarding requests
// to the underlying ServeMux.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

//
// HTTP Method Route Handlers
//
// All method handlers take a path, a name and a handler
// The name is used for URL reversing.
//
// Example:
//   router.Delete("/users/{id}", "delete-user", func(w http.ResponseWriter, r *http.Request) {
//       w.Write([]byte("User deleted"))
//   })

// Method-agnostic handler
// Handles request for any method
func (r *Router) Handle(path, name string, handler http.HandlerFunc) {
	r.Route(MethodAll, path, name, handler)
}

// Head adds a handler for HTTP HEAD requests to the given path.
func (r *Router) Head(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodHead, path, name, handler)
}

// Options adds a handler for HTTP OPTIONS requests to the given path.
func (r *Router) Options(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodOptions, path, name, handler)
}

// Delete adds a handler for HTTP DELETE requests to the given path.
func (r *Router) Delete(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodDelete, path, name, handler)
}

// Get adds a handler for HTTP GET requests to the given path.
func (r *Router) Get(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodGet, path, name, handler)
}

// Patch adds a handler for HTTP PATCH requests to the given path.
func (r *Router) Patch(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodPatch, path, name, handler)
}

// Post adds a handler for HTTP POST requests to the given path.
func (r *Router) Post(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodPost, path, name, handler)
}

// Put adds a handler for HTTP PUT requests to the given path.
func (r *Router) Put(path, name string, handler http.HandlerFunc) {
	r.Route(http.MethodPut, path, name, handler)
}

//
// Router utilities
//

// Include attaches another Router at the specified prefix, applying the parent router’s
// middleware to all routes in the sub-router.
//
// Example:
//
//	mainRouter := NewRouter("main")
//	apiRouter := NewRouter("api")
//	mainRouter.Include(apiRouter)
func (r *Router) Include(router *Router) {
	prefix := CleanPrefix(router.namespace)
	r.mux.Handle(
		prefix,
		http.StripPrefix(
			strings.TrimSuffix(prefix, "/"),
			ApplyMiddleware(router.ServeHTTP, r.middlewareFuncs...),
		),
	)
}

// Files serves static files (like images, CSS, or JavaScript) from a directory under
// the specified prefix. The router’s middleware is applied to all file requests.
//
// Example:
//
//	router.Files("/static/", "./public") // Serves files from ./public at /static/
func (r *Router) Files(prefix string, dirname string) {
	prefix = CleanPrefix(prefix)
	fs := http.FileServer(http.Dir(dirname))
	r.mux.Handle(
		prefix,
		http.StripPrefix(
			strings.TrimSuffix(prefix, "/"),
			ApplyMiddleware(fs.ServeHTTP, r.middlewareFuncs...),
		),
	)
}

// Register sets up a View to handle all supported HTTP methods (HEAD, OPTIONS, DELETE,
// GET, PATCH, POST, PUT) for the given path. The name is used for URL reversing.
//
// Example:
//
//	type MyView struct {}
//	func (v MyView) Dispatch(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Users list"))
//	}
//	router.Register("/users/", "users", MyView{})
func (r *Router) Register(path string, name string, view View) {
	r.Handle(path, name, view.Dispatch)
}

// Route registers the given method + path to a handler with middleware applied.
// Route adds a handler for a specific HTTP method (e.g., GET, POST) and path.
// The name is used for URL reversing.
//
// Panics if the path is empty or the HTTP method is invalid.
func (r *Router) Route(method, path, name string, handler http.HandlerFunc) {
	if path == "" {
		panic("router: path cannot be empty")
	}

	pattern := method + " " + path
	if method == MethodAll {
		pattern = path
	}

	r.mux.Handle(
		pattern,
		ApplyMiddleware(handler, r.middlewareFuncs...),
	)

	registryPath := path
	registryKey := name
	if r.namespace != "" {
		registryPath = "/" + r.namespace + path
		registryKey = r.namespace + ":" + name
	}
	routeRegistry[registryKey] = registryPath
}

//
// Utilities
//

// ApplyMiddleware wraps a handler with the provided middleware functions, applying
// them in reverse order (the last middleware runs first). This lets middleware
// process requests before or after the handler.
//
// Example:
//
//	handler = ApplyMiddleware(myHandler, loggerMiddleware, authMiddleware)
func ApplyMiddleware(handler http.HandlerFunc, middlewareFuncs ...Middleware) http.HandlerFunc {
	for i := len(middlewareFuncs) - 1; i >= 0; i-- {
		handler = middlewareFuncs[i](handler)
	}
	return handler
}

// CleanPrefix ensures a URL prefix starts and ends with a slash (e.g., "api" becomes
// "/api/"). This keeps routes consistent and avoids common URL mistakes.
func CleanPrefix(prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	return prefix
}

// ReverseUrl finds the full URL path for a named route in a namespace (e.g., "api:user-detail").
// It supports paths with placeholders like "/users/{name}/" by replacing them with values
// from the params string (e.g., "name=alice" ). Returns an error if the route isn’t found,
// a required parameter is missing, or an extra parameter is provided.
//
// Example:
//
//	url, err := ReverseUrl("api:user-detail", "name=alice"})
//	// Returns "/api/users/alice/" if registered as "/users/{name}/" under "/api/"
func ReverseUrl(name string, paramStr string) (string, error) {
	path, ok := routeRegistry[name]
	if !ok {
		return "", fmt.Errorf("no route registered for name %q", name)
	}

	matches := pathParamRegex.FindAllStringSubmatch(path, -1)
	required := make(map[string]bool, len(matches))
	for _, match := range matches {
		required[match[1]] = true
	}

	params, _ := url.ParseQuery(paramStr)
	for key, value := range params {
		if !required[key] {
			return "", fmt.Errorf("extra parameter %q not found in route %q", key, name)
		}

		placeholder := "{" + key + "}"
		path = strings.ReplaceAll(path, placeholder, value[0])
		delete(required, key)
	}

	if len(required) > 0 {
		for key := range required {
			return "", fmt.Errorf("missing required parameter %q for route %q", key, name)
		}
	}

	return path, nil
}

// ReverseUrlSimple is a convenience function to get the URL path for a route with no params.
//
// Example:
//
//	url := ReverseUrl("api:users-list")
//	// Returns "/users/" if registered
func ReverseUrlSimple(name string) (string, error) {
	return ReverseUrl(name, "")
}

//
// Type definitions
//

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
// single URL pattern. Implement this to define how a URL responds to different request types.
type View interface {
	Dispatch(w http.ResponseWriter, r *http.Request)
}
