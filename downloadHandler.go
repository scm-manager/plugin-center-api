package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
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
			protocol: "http", // TODO
		}
	}
}

func (u *UrlGenerator) DownloadUrl(plugin Plugin, version string) string {
	return fmt.Sprintf("%v://%v/api/v1/download/%v/%v", u.protocol, u.host, plugin.Name, version)
}

type DownloadHandler struct {
	plugins        map[string]Plugin
	downloadPlugin func(url string) (resp *http.Response, err error)
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
	handler := DownloadHandler{plugins: createMap(plugins), downloadPlugin: http.Get}
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
	plugin, ok := h.plugins[pluginName]
	if !ok {
		msg := fmt.Sprintf("no plugin found for name %s", pluginName)
		log.Println(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	authenticated := r.Context().Value("subject") != nil
	if plugin.RequiresAuthentication() && !authenticated {
		msg := fmt.Sprintf("plugin %s requires authentication", pluginName)
		log.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	pluginVersion := vars["version"]
	release := h.findRelease(plugin, pluginVersion)
	if release == nil {
		msg := fmt.Sprintf("no plugin found for name %s and version %s", pluginName, pluginVersion)
		log.Println(msg)
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	log.Println("found release for plugin", pluginName, "and version", pluginVersion, ":", release.Url)

	downloadCounter.WithLabelValues(
		pluginName,
		pluginVersion,
	).Inc()

	h.copyHttpStream(release, pluginName, pluginVersion, w)
}

func (h *DownloadHandler) copyHttpStream(release *Release, pluginName string, pluginVersion string, w http.ResponseWriter) {
	resp, err := h.downloadPlugin(release.Url)
	if err != nil {
		log.Println("error opening url for plugin", pluginName, "and version", pluginVersion, ":", release.Url, err)
		http.Error(w, "could not read plugin from target", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	w.Header().Add("Content-Disposition", `attachment; filename="`+pluginName+`.smp"`)
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Println("got an error copying download stream for url", release.Url, "after", written, "bytes:", err)
	}
}

func (h *DownloadHandler) findRelease(plugin Plugin, version string) *Release {
	for _, release := range plugin.Releases {
		if release.Version == version {
			return &release
		}
	}
	return nil
}
