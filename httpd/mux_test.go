package httpd_test

import (
	"fmt"
	"github.com/zenreach/redirector/httpd"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newMiddlewareRequest(t *testing.T, path string) *http.Request {
	url := fmt.Sprintf("http://example.com%s", path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}
	return req
}

func TestStatusMiddleware(t *testing.T) {
	t.Parallel()

	paths := []string{
		"ping",
		"_status",
	}

	for _, item := range paths {
		path := item
		t.Run(fmt.Sprintf("%s_no_slash", path), testMiddleware(path, fmt.Sprintf("/%s", path), true))
		t.Run(fmt.Sprintf("%s_slash", path), testMiddleware(path, fmt.Sprintf("/%s/", path), true))
		t.Run(fmt.Sprintf("%s_subpath", path), testMiddleware(path, fmt.Sprintf("/%s/subpath", path), false))
		t.Run(fmt.Sprintf("%s_users", path), testMiddleware(path, "/users", false))
	}
}

func testMiddleware(statusPath, urlPath string, status bool) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})
		middleware := httpd.NewStatusMiddleware(handler, nil, "", statusPath)

		w := httptest.NewRecorder()
		r := newMiddlewareRequest(t, urlPath)
		middleware.ServeHTTP(w, r)

		if status && w.Code != http.StatusOK {
			t.Errorf("status code incorrect: %d != %d", w.Code, http.StatusOK)
		} else if !status && w.Code != http.StatusNoContent {
			t.Errorf("handler code incorrect: %d != %d", w.Code, http.StatusNoContent)
		}
	}
}
