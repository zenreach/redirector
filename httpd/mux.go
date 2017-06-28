package httpd

import (
	"net/http"
	"strings"
)

type statusMux struct {
	path    string
	status  http.Handler
	handler http.Handler
}

func NewStatusMiddleware(handler http.Handler, checker Checker, version, path string) http.Handler {
	return &statusMux{
		path:    path,
		status:  NewStatusHandler(checker, version),
		handler: handler,
	}
}

// ServeHTTP serves an HTTP request.
func (h *statusMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := strings.Trim(h.path, "/")
	path := strings.Trim(r.URL.Path, "/")
	lPrefix := len(prefix)
	lPath := len(path)

	if lPath >= lPrefix && path[:lPrefix] == prefix {
		if lPath > lPrefix {
			h.handler.ServeHTTP(w, r)
		} else {
			h.status.ServeHTTP(w, r)
		}
	} else {
		h.handler.ServeHTTP(w, r)
	}
}
