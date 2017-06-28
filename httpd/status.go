package httpd

import (
	"encoding/json"
	"net/http"
)

// Status is the value returned by the status endpoint to indicate whether or
// not the service is live.
type Status string

const (
	Online Status = "online"
	Error  Status = "error"
)

// Checker is called by the status handler to check if the service is online.
// Checker returns the current service status as well as a list of errors
// currently being experienced by the service. It is possible for the service
// to return a list of errors while online.
type Checker func() (Status, []string)

type statusHandler struct {
	version string
	checker Checker
}

// NewStatusHandler creates an HTTP handler which services a service status
// page. The version should be the current version of the service. Checker is
// used to determine if the service is online or in an error state. If the
// version is an empty string it will be omitted from responses. If checker is
// nil then status will always report the service as online. The handler
// returns a 200 status code when online and a 503 when down.
func NewStatusHandler(checker Checker, version string) http.Handler {
	return &statusHandler{
		version: version,
		checker: checker,
	}
}

// ServeHTTP serves an HTTP request.
func (h *statusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	status := Online
	var errors []string

	if h.checker != nil {
		status, errors = h.checker()
	}

	type response struct {
		Status  Status   `json:"status"`
		Version string   `json:"version,omitempty"`
		Errors  []string `json:"errors,omitempty"`
	}

	bytes, err := json.Marshal(&response{
		Status:  status,
		Version: h.version,
		Errors:  errors,
	})
	if err == nil {
		if status == Online {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Write(bytes)
	} else {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
