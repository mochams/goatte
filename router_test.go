package goatte

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReverseUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}
	mainRouter := NewRouter("main")
	apiRouter := NewRouter("api")
	apiRouter.Get("/users/{name}/", "user-detail", handler)
	apiRouter.Post("/users/", "user-create", handler)
	mainRouter.Include(apiRouter)

	tests := []struct {
		title  string
		name   string
		params string
		want   string
		errMsg string
	}{
		{
			title:  "test parses path with params",
			name:   "api:user-detail",
			params: "name=alice",
			want:   "/api/users/alice/",
		},
		{
			title:  "test parses simple path",
			name:   "api:user-create",
			params: "",
			want:   "/api/users/",
		},
		{
			title:  "test returns error for unknown path",
			name:   "unknown",
			params: "",
			want:   "",
			errMsg: "no route registered for name \"unknown\"",
		},
		{
			title:  "test returns error for extra params",
			name:   "api:user-detail",
			params: "name=alice&id=1234",
			want:   "",
			errMsg: "extra parameter \"id\" not found in route \"api:user-detail\"",
		},
		{
			title:  "test returns error for missing params",
			name:   "api:user-detail",
			params: "",
			want:   "",
			errMsg: "missing required parameter \"name\" for route \"api:user-detail\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got, err := ReverseUrl(tt.name, tt.params)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if got != tt.want {
				t.Errorf("ReverseUrl(%q) Expected url `%q`, got `%q`", tt.name, tt.want, got)
			}

			if errMsg != tt.errMsg {
				t.Errorf("ReverseUrl(%q) Expected error message `%q`, got `%q`", tt.name, tt.errMsg, errMsg)
			}
		})
	}
}

func TestReverseSimpleUrl(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {}
	mainRouter := NewRouter("")
	mainRouter.Post("/users/", "user-create", handler)

	tests := []struct {
		title  string
		name   string
		params string
		want   string
		errMsg string
	}{
		{
			title: "test parses simple path",
			name:  ":user-create",
			want:  "/users/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got, _ := ReverseUrlSimple(tt.name)

			if got != tt.want {
				t.Errorf("ReverseUrl(%q) Expected url `%q`, got `%q`", tt.name, tt.want, got)
			}
		})
	}
}

func TestCleanPrefix(t *testing.T) {
	tests := []struct {
		title string
		path  string
		want  string
	}{
		{
			title: "test appends slash to prefix and suffix to path",
			path:  "user-create",
			want:  "/user-create/",
		},
		{
			title: "test appends suffix to path",
			path:  "/user-create",
			want:  "/user-create/",
		},
		{
			title: "test appends prefix to path",
			path:  "user-create/",
			want:  "/user-create/",
		},
		{
			title: "test does not modify path with prefix and suffix",
			path:  "/user-create/",
			want:  "/user-create/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := CleanPrefix(tt.path)

			if got != tt.want {
				t.Errorf("CleanPrefix(%q) Expected prefix `%q`, got `%q`", tt.path, tt.want, got)
			}
		})
	}
}

var testMiddleware = func(name string, tracker *[]string) func(http.HandlerFunc) http.HandlerFunc {
	middleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			*tracker = append(*tracker, "Enter-"+name)
			next.ServeHTTP(w, r)
			*tracker = append(*tracker, "Leave-"+name)
		}
	}
	return middleware
}

func TestApplyMiddleware_OrderOfExecution(t *testing.T) {

	tests := []struct {
		title      string
		middleware []Middleware
		expected   []string
	}{
		{
			title:      "test with multiple middleware",
			middleware: nil,
			expected: []string{
				"Enter-LOGGER",
				"Enter-AUTH",
				"Handler",
				"Leave-AUTH",
				"Leave-LOGGER",
			},
		},
		{
			title:      "test with one middleware",
			middleware: nil,
			expected: []string{
				"Enter-LOGGER",
				"Handler",
				"Leave-LOGGER",
			},
		},
		{
			title:      "test with no middleware",
			middleware: []Middleware{},
			expected: []string{
				"Handler",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			orderTracker := []string{}

			switch tt.title {
			case "test with multiple middleware":
				tt.middleware = []Middleware{
					testMiddleware("LOGGER", &orderTracker),
					testMiddleware("AUTH", &orderTracker),
				}
			case "test with one middleware":
				tt.middleware = []Middleware{
					testMiddleware("LOGGER", &orderTracker),
				}
			case "test with no middleware":
				tt.middleware = []Middleware{}
			}

			handler := func(w http.ResponseWriter, r *http.Request) {
				orderTracker = append(orderTracker, "Handler")
				w.WriteHeader(http.StatusOK)
			}

			wrapped := ApplyMiddleware(handler, tt.middleware...)
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			wrapped(w, req)

			if len(orderTracker) != len(tt.expected) {
				t.Fatalf("expected %d events, got %d", len(tt.expected), len(orderTracker))
			}

			for i, want := range tt.expected {
				if got := orderTracker[i]; got != want {
					t.Errorf("order[%d]: expected %q, got %q", i, want, got)
				}
			}

			clear((orderTracker))
		})
	}
}
