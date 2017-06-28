package httpd_test

import (
	"bytes"
	"encoding/json"
	"github.com/zenreach/redirector/httpd"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func newStatusRequest(t *testing.T, method string, payload []byte) *http.Request {
	var rdr io.Reader
	if payload != nil {
		rdr = bytes.NewBuffer(payload)
	}

	url := "http://example.com/_status/"
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("failed to create request: %s", err)
	}
	return req
}

func TestStatusGet(t *testing.T) {
	t.Parallel()

	type Test struct {
		Name      string
		Status    httpd.Status
		Version   string
		Errors    []string
		NoChecker bool
		Want      string
	}

	tests := []Test{
		{
			Name:    "online_version",
			Status:  httpd.Online,
			Version: "1.0.0",
			Want:    `{"status":"online","version":"1.0.0"}`,
		},
		{
			Name:   "online",
			Status: httpd.Online,
			Want:   `{"status":"online"}`,
		},
		{
			Name:   "online_version_errors",
			Status: httpd.Online,
			Errors: []string{
				"service degraded",
			},
			Want: `{"status":"online","errors":["service degraded"]}`,
		},
		{
			Name:    "error_version_errors",
			Status:  httpd.Error,
			Version: "3.2.1",
			Errors: []string{
				"service unavailable",
			},
			Want: `{"status":"error","version":"3.2.1","errors":["service unavailable"]}`,
		},
		{
			Name:    "error_version",
			Status:  httpd.Error,
			Version: "1.2.0",
			Want:    `{"status":"error","version":"1.2.0"}`,
		},
		{
			Name:   "error_errors",
			Status: httpd.Error,
			Errors: []string{
				"service unavailable",
			},
			Want: `{"status":"error","errors":["service unavailable"]}`,
		},
		{
			Name:      "online_nochecker",
			NoChecker: true,
			Want:      `{"status":"online"}`,
		},
		{
			Name:      "error_nochecker_version",
			Version:   "4.1.1",
			NoChecker: true,
			Want:      `{"status":"online","version":"4.1.1"}`,
		},
	}

	for _, item := range tests {
		test := item
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			want := make(map[string]interface{})
			err := json.Unmarshal([]byte(test.Want), &want)
			if err != nil {
				t.Fatal(err)
			}

			var checker httpd.Checker
			if !test.NoChecker {
				checker = func() (httpd.Status, []string) {
					return test.Status, test.Errors
				}
			}
			handler := httpd.NewStatusHandler(checker, test.Version)

			w := httptest.NewRecorder()
			r := newStatusRequest(t, "GET", nil)
			handler.ServeHTTP(w, r)

			if test.Status == httpd.Online && w.Code != http.StatusOK {
				t.Errorf("online status code incorrect: %d != %d", w.Code, http.StatusOK)
			} else if test.Status == httpd.Error && w.Code != http.StatusServiceUnavailable {
				t.Error("error status code incorrect: %d != %d", w.Code, http.StatusServiceUnavailable)
			}

			body := w.Body.Bytes()
			have := make(map[string]interface{})
			err = json.Unmarshal(body, &have)
			if err != nil {
				t.Error("failed to decode body: %s", err)
				t.Error("body: %s", body)
				return
			}

			if !reflect.DeepEqual(have, want) {
				t.Error("got unexpected response:")
				t.Errorf("  have: %+v", have)
				t.Errorf("  want: %+v", want)
			}
		})
	}
}

func TestStatusMethodNotAllowed(t *testing.T) {
	t.Parallel()

	type Test struct {
		Method  string
		Payload string
	}

	tests := []Test{
		{
			Method:  "POST",
			Payload: "test post",
		},
		{
			Method:  "PUT",
			Payload: "test put",
		},
		{
			Method: "DELETE",
		},
	}

	for _, item := range tests {
		test := item
		t.Run(item.Method, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := newStatusRequest(t, test.Method, []byte(test.Payload))
			handler := httpd.NewStatusHandler(nil, "")
			handler.ServeHTTP(w, r)

			if w.Code != http.StatusMethodNotAllowed {
				t.Error("status code incorrect: %d != %d", w.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
