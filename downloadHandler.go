package main

import (
	"fmt"
	"net/http"
)

type UrlGenerator struct {
	host     string
	protocol string
}

func NewUrlGenerator(r http.Request) UrlGenerator {
	forwardedHost := r.Header.Get("X-Forwarded-Host")
	forwardedProto := r.Header.Get("X-Forwarded-Proto")
	if forwardedHost != "" && forwardedProto != "" {
		return UrlGenerator{
			host:     forwardedHost,
			protocol: forwardedProto,
		}
	} else {
		return UrlGenerator{
			host:     r.Host,
			protocol: r.Proto,
		}
	}
}

func (u *UrlGenerator) DownloadUrl(plugin Plugin, version string) string {
	return fmt.Sprintf("%v://%v/api/v1/download/%v/%v", u.protocol, u.host, plugin.Name, version)
}

func NewDownloadHandler(plugins []Plugin) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
