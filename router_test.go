package goatte

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
			name:  "user-create",
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

func TestApplyMiddleware(t *testing.T) {

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

func TestNewRouter(t *testing.T) {
	tests := []struct {
		title     string
		namespace string
		params    string
		want      string
	}{
		{
			title:     "test get new router",
			namespace: "main",
		},
		{
			title:     "test get new router",
			namespace: "api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := NewRouter(tt.namespace)

			if got.namespace != tt.namespace {
				t.Errorf("NewRouter(%q) Expected namespace `%q`, got `%q`", tt.namespace, tt.namespace, got.namespace)
			}
		})
	}
}

func TestRouteHandlers_GET(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test GET returns 200",
			method:   "GET",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Get("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_POST(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test POST returns 200",
			method:   "POST",
			expected: 200,
		},
		{
			title:    "test PUT returns 405",
			method:   "GET",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Post("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_PUT(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test PUT returns 200",
			method:   "PUT",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Put("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_PATCH(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test PATCH returns 200",
			method:   "PATCH",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Patch("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_DELETE(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test DELETE returns 200",
			method:   "DELETE",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Delete("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_OPTIONS(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test OPTIONS returns 200",
			method:   "OPTIONS",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Options("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouteHandlers_HEAD(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test HEAD returns 200",
			method:   "HEAD",
			expected: 200,
		},
		{
			title:    "test POST returns 405",
			method:   "POST",
			expected: 405,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Head("/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if res.StatusCode == http.StatusAccepted && !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

// Test View
type TestView struct {
	called bool
}

// Implementing the Dispatch method
func (v *TestView) Dispatch(w http.ResponseWriter, r *http.Request) {
	v.called = true
	w.WriteHeader(http.StatusOK)
}

func TestRouter_Register(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test HEAD returns 200",
			method:   "HEAD",
			expected: 200,
		},
		{
			title:    "test OPTIONS returns 200",
			method:   "OPTIONS",
			expected: 200,
		},
		{
			title:    "test GET returns 200",
			method:   "GET",
			expected: 200,
		},
		{
			title:    "test POST returns 200",
			method:   "POST",
			expected: 200,
		},
		{
			title:    "test PUT returns 200",
			method:   "PUT",
			expected: 200,
		},
		{
			title:    "test PATCH returns 200",
			method:   "PATCH",
			expected: 200,
		},
		{
			title:    "test DELETE returns 200",
			method:   "DELETE",
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")

			view := &TestView{called: false}

			router.Register("/", "test-detail", view)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if !view.called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouter_Files(t *testing.T) {
	tmpDir := t.TempDir()
	testFilePath := filepath.Join(tmpDir, "testfile.css")
	err := os.WriteFile(testFilePath, []byte("body { background-color: red; }"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	router := NewRouter("test")
	router.Files("/static/", tmpDir)

	req := httptest.NewRequest("GET", "/static/testfile.css", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	res := w.Result()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", res.StatusCode)
	}

	contentType, expected := res.Header.Get("Content-Type"), "text/css; charset=utf-8"
	if contentType != expected {
		t.Errorf("expected Content-Type to contain '%s', got '%s'", expected, contentType)
	}
}

func TestRouter_Route(t *testing.T) {
	tests := []struct {
		title    string
		method   string
		expected int
	}{
		{
			title:    "test HEAD returns 200",
			method:   "HEAD",
			expected: 200,
		},
		{
			title:    "test OPTIONS returns 200",
			method:   "OPTIONS",
			expected: 200,
		},
		{
			title:    "test GET returns 200",
			method:   "GET",
			expected: 200,
		},
		{
			title:    "test POST returns 200",
			method:   "POST",
			expected: 200,
		},
		{
			title:    "test PUT returns 200",
			method:   "PUT",
			expected: 200,
		},
		{
			title:    "test PATCH returns 200",
			method:   "PATCH",
			expected: 200,
		},
		{
			title:    "test DELETE returns 200",
			method:   "DELETE",
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Route(tt.method, "/", "test-detail", handler)
			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}

func TestRouter_Route_Panic(t *testing.T) {
	router := NewRouter("test")

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("Expected panic but did not occur")
		}

		expectedMessage := "router: path cannot be empty"
		if r != expectedMessage {
			t.Errorf("Expected panic message '%s', but got '%v'", expectedMessage, r)
		}
	}()

	router.Route("GET", "", "test-detail", handler)
}

func TestRouter_Handle(t *testing.T) {
	tests := []struct {
		title    string
		path     string
		method   string
		expected int
	}{
		{
			title:    "test handle /",
			path:     "/",
			method:   "POST",
			expected: 200,
		},
		{
			title:    "test handle /test",
			path:     "/test",
			method:   "POST",
			expected: 200,
		},
		{
			title:    "test handle /panel",
			path:     "/panel",
			method:   "POST",
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			router := NewRouter("test")
			called := false

			handler := func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}

			router.Handle(tt.path, "test-detail", handler)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			res := w.Result()

			if !called {
				t.Error("handler was not called")
			}

			if res.StatusCode != tt.expected {
				t.Errorf("expected 200 OK, got %d", res.StatusCode)
			}
		})
	}
}
