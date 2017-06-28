package httpd_test

import (
	"fmt"
	"github.com/zenreach/redirector/httpd"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newRedirectRequest(t *testing.T, url, host, proto string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}
	if host != "" {
		req.Header.Set("Host", host)
	}
	if proto != "" {
		req.Header.Set("X-Forwarded-Proto", proto)
	}
	return req
}

func TestHTTPSUpgrade(t *testing.T) {
	t.Parallel()

	type testEntry struct {
		URL      string
		Proto    string
		Status   int
		Location string
	}

	tests := []testEntry{
		// HTTPS redirect with no header
		{URL: "http://example.com", Status: 307, Location: "https://example.com"},
		{URL: "http://example.com/", Status: 307, Location: "https://example.com/"},
		{URL: "http://example.com/no/slash", Status: 307, Location: "https://example.com/no/slash"},
		{URL: "http://example.com/trailing/slash/", Status: 307, Location: "https://example.com/trailing/slash/"},
		{URL: "http://127.0.0.1", Status: 307, Location: "https://127.0.0.1"},
		// no redirect with no header
		{URL: "https://example.com", Status: 404},
		{URL: "https://example.com/", Status: 404},
		{URL: "https://127.0.0.1", Status: 404},
		// HTTPS redirect with header
		{URL: "http://example.com", Proto: "http", Status: 307, Location: "https://example.com"},
		{URL: "http://example.com", Proto: "HTTP", Status: 307, Location: "https://example.com"},
		// no redirect with header
		{URL: "http://example.com", Proto: "https", Status: 404},
		{URL: "http://example.com", Proto: "HTTPS", Status: 404},
		// with port
		{URL: "http://example.com:8080", Status: 307, Location: "https://example.com:8080"},
		{URL: "http://127.0.0.1:8080", Status: 307, Location: "https://127.0.0.1:8080"},
		{URL: "https://example.com:8080", Status: 404},
		{URL: "https://127.0.0.1:8080", Status: 404},
		{URL: "http://example.com:8080", Proto: "http", Status: 307, Location: "https://example.com:8080"},
		{URL: "http://127.0.0.1:8080", Proto: "http", Status: 307, Location: "https://127.0.0.1:8080"},
		// strange situations
		{URL: "http://127.0.0.1:80", Status: 307, Location: "https://127.0.0.1:80"},
	}

	handler := httpd.NewRedirectHandler(true, false)

	for n, item := range tests {
		test := item
		t.Run(fmt.Sprintf("Test%d", n), func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := newRedirectRequest(t, test.URL, "", test.Proto)
			handler.ServeHTTP(w, r)

			if w.Code != test.Status {
				t.Errorf("wrong status code: %d != %d", w.Code, test.Status)
			}
			location := w.HeaderMap.Get("Location")
			if test.Status == 307 && location != test.Location {
				t.Errorf("wrong Location: %s != %s", location, test.Location)
			}
		})
	}
}

func TestWWWUpgrade(t *testing.T) {
	t.Parallel()

	type testEntry struct {
		URL      string
		Host     string
		Status   int
		Location string
	}

	tests := []testEntry{
		// redirect with no header
		{URL: "http://example.com", Status: 307, Location: "http://www.example.com"},
		{URL: "http://example.com/", Status: 307, Location: "http://www.example.com/"},
		// no redirect with no header
		{URL: "http://www.example.com", Status: 404},
		{URL: "http://www.example.com/", Status: 404},
		// redirect with header
		{URL: "http://localhost", Host: "example.com", Status: 307, Location: "http://www.example.com"},
		{URL: "http://localhost/", Host: "example.com", Status: 307, Location: "http://www.example.com/"},
		// no redirect with header
		{URL: "http://localhost", Host: "www.example.com", Status: 404},
		{URL: "http://localhost/", Host: "www.example.com", Status: 404},
		// https not stripped
		{URL: "https://example.com", Status: 307, Location: "https://www.example.com"},
		{URL: "https://example.com/", Status: 307, Location: "https://www.example.com/"},
		// ip address not prepended
		{URL: "http://127.0.0.1", Status: 404},
		{URL: "http://127.0.0.1/", Status: 404},
		{URL: "https://127.0.0.1", Status: 404},
		{URL: "https://127.0.0.1/", Status: 404},
		// with port
		{URL: "http://example.com:8080", Status: 307, Location: "http://www.example.com:8080"},
		{URL: "http://www.example.com:8080", Status: 404},
		{URL: "http://localhost:8080", Host: "example.com", Status: 307, Location: "http://www.example.com:8080"},
		{URL: "http://localhost:8080", Host: "www.example.com", Status: 404},
		{URL: "https://example.com:8080", Status: 307, Location: "https://www.example.com:8080"},
		{URL: "http://127.0.0.1:8080", Status: 404},
	}

	handler := httpd.NewRedirectHandler(false, true)

	for n, item := range tests {
		test := item
		t.Run(fmt.Sprintf("Test%d", n), func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := newRedirectRequest(t, test.URL, test.Host, "")
			handler.ServeHTTP(w, r)

			if w.Code != test.Status {
				t.Errorf("wrong status code: %d != %d", w.Code, test.Status)
			}
			location := w.HeaderMap.Get("Location")
			if test.Status == 307 && location != test.Location {
				t.Errorf("wrong Location: %s != %s", location, test.Location)
			}
		})
	}
}

func TestBothUpgrades(t *testing.T) {
	t.Parallel()

	type testEntry struct {
		URL      string
		Host     string
		Proto    string
		Status   int
		Location string
	}

	tests := []testEntry{
		// no redirect with no headers
		{URL: "https://www.example.com", Status: 404},
		{URL: "https://www.example.com/", Status: 404},
		// no redirect with headers
		{URL: "http://127.0.0.1", Proto: "https", Host: "www.example.com", Status: 404},
		// redirect with no headers
		{URL: "http://example.com", Status: 307, Location: "https://www.example.com"},
		{URL: "https://example.com", Status: 307, Location: "https://www.example.com"},
		{URL: "http://www.example.com", Status: 307, Location: "https://www.example.com"},
		// redirect with host header
		{URL: "http://127.0.0.1", Host: "example.com", Status: 307, Location: "https://www.example.com"},
		{URL: "https://127.0.0.1", Host: "example.com", Status: 307, Location: "https://www.example.com"},
		// redirect with proto header
		{URL: "http://example.com", Proto: "http", Status: 307, Location: "https://www.example.com"},
		{URL: "http://www.example.com", Proto: "http", Status: 307, Location: "https://www.example.com"},
		// redirect with headers
		{URL: "http://127.0.0.1", Proto: "http", Host: "example.com", Status: 307, Location: "https://www.example.com"},
		{URL: "http://example.com", Proto: "http", Host: "example.com", Status: 307, Location: "https://www.example.com"},
	}

	handler := httpd.NewRedirectHandler(true, true)

	for n, item := range tests {
		test := item
		t.Run(fmt.Sprintf("Test%d", n), func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := newRedirectRequest(t, test.URL, test.Host, test.Proto)
			handler.ServeHTTP(w, r)

			if w.Code != test.Status {
				t.Errorf("wrong status code: %d != %d", w.Code, test.Status)
			}
			location := w.HeaderMap.Get("Location")
			if test.Status == 307 && location != test.Location {
				t.Errorf("wrong Location: %s != %s", location, test.Location)
			}
		})
	}
}
