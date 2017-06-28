package main

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/zenreach/redirector/httpd"
	"net/http"
	"os"
)

type config struct {
	UpgradeSSL bool   `env:"SSL"`
	UpgradeWWW bool   `env:"WWW"`
	Listen     string `env:"LISTEN"`
	Status     string `env:"STATUS"`
}

func main() {
	cfg := &config{
		UpgradeSSL: true,
		UpgradeWWW: false,
		Listen:     "0.0.0.0:80",
		Status:     "_status",
	}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse env vars: %s\n", err)
		os.Exit(1)
	}

	redirector := httpd.NewRedirectHandler(cfg.UpgradeSSL, cfg.UpgradeWWW)
	status := httpd.NewStatusMiddleware(redirector, nil, "", cfg.Status)
	if err := http.ListenAndServe(cfg.Listen, status); err != nil {
		fmt.Fprintf(os.Stderr, "http server failed: %s\n", err)
		os.Exit(1)
	}
}
