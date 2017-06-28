package httpd

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type redirectHandler struct {
	upgradeSSL bool
	upgradeWWW bool
}

// NewRedirectHandler creates a new redirect handler. The `upgradeSSL` enables
// redirection from http to https. The `upgradeWWW` options enables redirection
// from `example.com` to `www.example.com`.
func NewRedirectHandler(upgradeSSL, upgradeWWW bool) http.Handler {
	return &redirectHandler{
		upgradeSSL: upgradeSSL,
		upgradeWWW: upgradeWWW,
	}
}

// ServeHTTP serves an HTTP request.
func (h *redirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redirect := false
	location := copyURL(r.URL)

	// favor X-Forwarded-Proto over URL
	proto := strings.ToLower(r.Header.Get("X-Forwarded-Proto"))
	if proto == "" {
		proto = strings.ToLower(location.Scheme)
	}

	// parse the host:port
	host, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		host = r.Host
	}
	hostHeader := r.Header.Get("Host")
	if hostHeader != "" {
		host = hostHeader
	}

	// upgrade to https if enabled
	if h.upgradeSSL && proto != "https" {
		proto = "https"
		redirect = true
	}

	// upgrade to www.* if enabled
	if h.upgradeWWW && !strings.HasPrefix(host, "www.") && !isIPAddress(host) {
		host = fmt.Sprintf("www.%s", host)
		redirect = true
	}

	if redirect {
		if port != "" {
			host = net.JoinHostPort(host, port)
		}
		location = copyURL(r.URL)
		location.Scheme = proto
		location.Host = host
		http.Redirect(w, r, location.String(), http.StatusTemporaryRedirect)
	} else {
		http.NotFound(w, r)
	}
}

// copyURL copies a given URL.
func copyURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	cp := *u
	if u.User != nil {
		user := *u.User
		cp.User = &user
	}
	return &cp
}

// iSIPAddress returns true if the hostname is an IP address.
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}
