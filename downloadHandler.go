package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	plugins map[string]map[string]Release
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

func createMap(plugins []Plugin) map[string]map[string]Release {
	pluginMap := make(map[string]map[string]Release)
	for _, plugin := range plugins {
		releaseMap := make(map[string]Release)
		for _, release := range plugin.Releases {
			releaseMap[release.Version] = release
		}
		pluginMap[plugin.Name] = releaseMap
	}
	return pluginMap
}

func (h *DownloadHandler) handle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pluginName := vars["plugin"]
	pluginVersion := vars["version"]

	release := h.findRelease(pluginName, pluginVersion)

	if release.Version == "" {
		log.Println("no release found for plugin", pluginName, "and version", pluginVersion)
		w.WriteHeader(404)
		return
	}
	log.Println("found release found for plugin", pluginName, "and version", pluginVersion, ":", release.Url)

	downloadCounter.WithLabelValues(
		pluginName,
		pluginVersion,
	).Inc()

	releaseUrl, err := url.ParseRequestURI(release.Url)

	if err != nil {
		log.Println("could not parse url for release:", err)
		w.WriteHeader(500)
		w.Write([]byte("illegal url for plugin found"))
		return
	}

	director := func(req *http.Request) {
		req.URL.Scheme = releaseUrl.Scheme
		req.URL.Host = releaseUrl.Host
		req.URL.Path = releaseUrl.Path
		req.URL.RawQuery = releaseUrl.RawQuery
		req.Host = releaseUrl.Host

		log.Println("redirecting download from", r.URL.String(), "to", req.URL.String())
	}

	proxy := httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}

func (h *DownloadHandler) findRelease(name string, version string) Release {
	releases := h.plugins[name]
	return releases[version]
}
