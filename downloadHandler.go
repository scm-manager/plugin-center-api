package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
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
			protocol: "https",
		}
	}
}

func (u *UrlGenerator) DownloadUrl(plugin Plugin, version string) string {
	return fmt.Sprintf("%v://%v/api/v1/download/%v/%v", u.protocol, u.host, plugin.Name, version)
}

type DownloadHandler struct {
	plugins map[string]Plugin
}

var (
	downloadCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "scm_plugin_center_download_requests",
		Help: "Total number of downloads",
	}, []string{
		"plugin", "version",
	})
)

func NewDownloadHandler(plugins []Plugin) http.HandlerFunc {
	handler := DownloadHandler{plugins: createMap(plugins)}
	return handler.handle
}

func createMap(plugins []Plugin) map[string]Plugin {
	m := make(map[string]Plugin)
	for _, plugin := range plugins {
		m[plugin.Name] = plugin
	}
	return m
}

func (h *DownloadHandler) handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginName := vars["plugin"]
	pluginVersion := vars["version"]

	release := h.findRelease(pluginName, pluginVersion)

	if release == nil {
		log.Println("no plugin found for name", pluginName, "and version", pluginVersion)
		w.WriteHeader(404)
		return
	}

	downloadCounter.WithLabelValues(
		pluginName,
		pluginVersion,
	).Inc()

	http.Redirect(w, r, release.Url, http.StatusSeeOther)
}

func (h *DownloadHandler) findRelease(name string, version string) *Release {
	plugin := h.plugins[name]
	for _, release := range plugin.Releases {
		if release.Version == version {
			return &release
		}
	}
	return nil
}
